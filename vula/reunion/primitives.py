from hashlib import blake2b
from hkdf import Hkdf
from sibc.csidh import CSIDH
from monocypher.secret import SecretBox, CryptoError
from monocypher.public import PublicKey, PrivateKey, Box
from monocypher.pwhash import argon2i
from monocypher._monocypher import lib, ffi
from cryptography.hazmat.primitives.ciphers import Cipher, algorithms, modes


def Hash(msg):
    return blake2b(msg).digest()[:32]


def x25519(sk, pk):
    return Box(sk, pk).shared_key()


def aead_encrypt(key, msg, ad):
    """
    Where the paper specifies an AEAD, we're using SecretBox's nonce as the AD.
    This is very much *not* the same as an AEAD, as the nonce should not be
    reused with the same key whereas the AD may be. In our case, we are never
    encrypting multiple messages with the same key, so this should be OK?

    We should probably replace this with an actual AEAD construction :)
    """
    ct = SecretBox(key).encrypt(msg, nonce=Hash(ad)[:24])
    return ct.detached_mac + ct.detached_ciphertext


def aead_decrypt(key, msg, ad):
    mac, ct = msg[:16], msg[16:]
    try:
        res = SecretBox(key).decrypt_raw(ct, Hash(ad)[:24], mac)
    except CryptoError as ex:
        # this happens if the mac failed. callers should check for None and
        # behave accordingly.
        res = None
    except Exception as ex:
        # we can arrive here if, for instance, msg is less than 16 bytes
        # (0-length ct and 16 byte mac is the shortest msg allows). currently
        # we drop to pdb in this case :)
        res = None
        exx = ex
        import pdb

        pdb.set_trace()
    return res


def elligator(curve, tweak):
    curve = ffi.from_buffer("uint8_t[32]", curve)
    hidden = ffi.new("uint8_t[32]")
    lib.crypto_curve_to_hidden(hidden, curve, tweak)
    return bytes(hidden)


def unelligator(hidden):
    hidden = ffi.from_buffer("uint8_t[32]", hidden)
    curve = ffi.new("uint8_t[32]")
    lib.crypto_hidden_to_curve(curve, hidden)
    return bytes(curve)


def generate_elligatorable_sk():
    while True:
        sk = PrivateKey.generate()
        pk = sk.public_key.encode()
        if pk == unelligator(elligator(pk, 0)):
            break
    return sk.encode()


def prp_encrypt(key, msg):
    # 32-byte block cipher rijndael-enc(key, plaintext)
    # note: should actually be rijandael with a 256 bit block size, but here it
    # is two blocks of AES (128 bit block size) for now.
    cipher = Cipher(algorithms.AES(key), modes.ECB())
    encryptor = cipher.encryptor()
    return encryptor.update(msg) + encryptor.finalize()


def prp_decrypt(key, ct):
    # 32-byte block cipher rijndael-dec(key, ciphertext),
    # note: should actually be rijandael with a 256 bit block size, but here it
    # is two blocks of AES (128 bit block size) for now.
    cipher = Cipher(algorithms.AES(key), modes.ECB())
    decryptor = cipher.decryptor()
    return decryptor.update(ct) + decryptor.finalize()


csidh_parameters = dict(
    curvemodel="montgomery",
    prime="p512",
    formula="hvelu",
    style="df",
    exponent=10,
    tuned=True,
    uninitialized=False,
    multievaluation=False,
    verbose=False,
)

csidh = CSIDH(**csidh_parameters)
