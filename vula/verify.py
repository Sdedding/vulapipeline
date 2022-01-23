"""
 vula-verify is a stateless program that generates and reads QR codes to
 verify a peer or a set of vula peers.  This program's output is intended to
 be fed into the vula-organize daemon. It should authenticate an already
 known peer or peers and sets a bit to keep state that it has verified them.
 The QR code also includes a PSK and later will use CSIDH to automatically set
 a PSK on a pair-wise basis.

 The output of this program may be written to a pipe, a log file, a unix
 socket, or any other place. It should run with the lowest possible privileges
 possible. The output is not filtered and so adversaries may attempt to inject
 unreasonable data. Care should be taken that the data should only be used
 after it has been verified.
"""
from cryptography.hazmat.primitives.ciphers.aead import ChaCha20Poly1305
from sibc.csidh import CSIDH

from vula.common import b64_bytes

try:
    import cv2
    from pyzbar.pyzbar import decode
except ImportError:
    zbar = None
    cv2 = None
    decode = None

try:
    import qrcode
except ImportError:
    qrcode = None

import json
import yaml
import click
from click.exceptions import Exit
import pydbus

from .constants import (
    _ORGANIZE_DBUS_NAME,
    _ORGANIZE_DBUS_PATH,
)
from .notclick import DualUse, green, bold, blue
from .peer import Descriptor
from .engine import Result


@DualUse.object(
    short_help="Verify and share peer verification information",
    invoke_without_command=False,
)
@click.pass_context
class VerifyCommands(object):
    def __init__(self, ctx):
        organize = ctx.meta.get('Organize', {}).get('magic_instance')

        if not organize:
            bus = pydbus.SystemBus()
            organize = bus.get(_ORGANIZE_DBUS_NAME, _ORGANIZE_DBUS_PATH)

        self.organize = organize
        self.my_descriptors = {
            ip: Descriptor(d)
            for ip, d in json.loads(
                self.organize.our_latest_descriptors()
            ).items()
        }
        (self.vk,) = set(d.vk for d in self.my_descriptors.values())

    @DualUse.method()
    def my_vk(self):
        click.echo(green(bold("Your VK is: ")) + str(self.vk))
        qr = qrcode.QRCode()
        qr.add_data(data="local.vula:vk:" + str(self.vk))
        qr.print_ascii()

    @DualUse.method()
    def my_descriptor(self):
        for ip, desc in self.my_descriptors.items():
            click.echo(green(bold("Descriptor for {}: ".format(ip))))
            qr = qrcode.QRCode()
            qr.add_data(data="local.vula:desc:" + str(desc))
            qr.print_ascii()
            click.echo(repr(str(desc)))

    @DualUse.method()
    @click.argument('name', type=str)
    def against(self, name):
        """
        Traceback (most recent call last):
          File "/usr/lib/python3/dist-packages/vula/verify.py",
          line 114, in against
            ss = self.organize.dh(str(pk))
          File "/usr/lib/python3/dist-packages/pydbus/proxy_method.py",
          line 72, in __call__
            ret = instance._bus.con.call_sync(
        gi.repository.GLib.Error:
        g-io-error-quark: GDBus.Error:unknown.
        AssertionError: non-supersingular input curve (36)
        """
        click.echo(green(bold("Verify against for {}").format(name)))
        pk = b64_bytes(self.organize.peer_pk(name))
        click.echo(blue(bold("pk = {}").format(str(pk))))
        sk = self.organize.generate_or_read_sk()
        click.echo(blue(bold("sk = {}").format(str(sk))))
        # ss = CSIDH.derive(sk=sk, pk=pk)
        ss = self.organize.dh(str(pk))

        self_name = self.my_descriptor.hostname
        message = self.vk
        cipher = ChaCha20Poly1305.encrypt(nonce=bytes(ss), data=bytes(message))
        qr = qrcode.QRCode()
        data = "local.vula:aead:" + str(cipher) + ":" + str(self_name)
        qr.add_data(data=data)
        qr.print_ascii()
        click.echo(green(bold("data: {}").format(data)))
        pass

    @DualUse.method()
    @click.option('-w', '--width', default=640, show_default=True)
    @click.option('-h', '--height', default=480, show_default=True)
    @click.option('-c', '--camera', default=0, show_default=True)
    @click.option(
        '-d', '--debug', default=False, is_flag=True, show_default=True
    )
    @click.argument('hostname', type=str, required=True)
    def scan(self, width, height, camera, hostname, debug):
        """
        We expect a string object that roughly looks like the following three
        things:


            local.vula:desc:<descriptor base64 representation>
            local.vula:vk:<vk base64 representation>
            local.vula:aead:<aead ciphertext base64 representation>

        The first part of the string is conformant to RFC 1738, Section 2.1.
        which describes "The main parts of URLs".
        """
        res = None
        done = False
        data = None
        v = cv2.VideoCapture(camera, cv2.CAP_V4L)
        v.set(cv2.CAP_PROP_FRAME_WIDTH, width)
        v.set(cv2.CAP_PROP_FRAME_HEIGHT, height)
        while v.isOpened():
            ret, img = v.read()
            if not ret:
                break
            cv2.imshow("scan vula qrcode", img)
            k = cv2.waitKey(1)
            if k == ord('q'):
                break
            res = decode(img)
            if res:
                if debug:
                    click.echo(res)
                for result in res:
                    if result.type == 'QRCODE':
                        if result.data.decode('utf-8').startswith(
                            'local.vula:'
                        ):
                            data = result.data.decode('utf-8')
                            done = True
                            break
            if done:
                break
        if not data:
            exit(1)
        data = data.split(':', 1)[1]
        sub_type, data = data.split(':', 1)
        if sub_type == "desc":
            res = self.organize.process_descriptor_string(data)
            res = Result(yaml.safe_load(res))
            click.echo(res)
        elif sub_type == "vk":
            vk = self.organize.get_vk_by_name(hostname)
            if vk == data:
                res = self.organize.verify_and_pin_peer(vk, hostname)
                res = Result(yaml.safe_load(res))
                click.echo(res)
                if res.error is not None:
                    raise Exception(res.error)
            else:
                click.echo("keys are for the wrong DeLorean")
                raise Exit(1)
        elif sub_type == "aead":
            self.handle_aead(data)
        else:
            click.echo("unknown qrcode subtype")
            raise Exit(1)
        if res.error is not None:
            click.echo(res)
            raise Exit(1)
        else:
            if debug:
                click.echo(res)
            raise Exit(0)

    def handle_aead(self, data):
        cipher, hostname = data.split(":")
        pk = self.organize.peer_pk(hostname)
        click.echo(blue(bold("pk = {}").format(pk)))

        sk = self.organize.generate_or_read_sk()
        click.echo(blue(bold("sk = {}").format(str(sk))))

        ss = CSIDH.dh(sk, pk)
        message = ChaCha20Poly1305.decrypt(nonce=bytes(ss), data=bytes(cipher))
        if message == self.organize.get_vk_by_name(hostname):
            # self.organize.verify_and_pin_peer(message, hostname)
            click.echo(
                green(bold("VERIFIED {} with vk {}").format(hostname, message))
            )
        else:
            click.echo("wrong")
            raise Exit(1)


main = VerifyCommands.cli
if __name__ == "__main__":
    main()
