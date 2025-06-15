# Current State of Vula on macOS 2025

June 2025 — authors **Bohdan Potuzhnyi** and **Vlada Svirsh**

[0. Prerequisites](#prereq)
[1. Installation](#installation)
[2. Tests](#tests)
[3. Running?](#debugging)
[4. Conclusion](#conclusion)

This is a second overview of the current state of Vula on macOS. 
Its intention is to highlight the areas that still need work in order
to get Vula **working** on macOS.

<a name="prereq"></a>

## 0. Authors' Word

As discussed in the earlier paper from 2022, installation **is** possible. If you run `pytest`, expect eight
errors, and if you try to launch Vula, it will crash from the **beginning**. Having obtained the same results during
initial research, it became clear that we at least needed to find out why the tests were failing and come
up with a strategy.

After launching the tests, we found that 4 tests were failing because of an issue with `IPRoute` and 4 because of an
issue in `PyRoute2WireGuard`. Both of these classes come from Vula's dependency `pyroute2`. Removing this dependency
would be like shooting the project in the foot, so a simpler approach—changing the dependent library—was chosen.
After checking the package, we identified that it already detects the system and assigns the correct implementation.
The only thing left was to look at the existing implementations for Windows and other systems and recreate something
similar for Darwin. After this change, we were able to remove four failing tests related to calls to `pyroute2`.

The change to the WireGuard-related file was more of a challenge, as it would require creating mechanics
similar to those in `IPRoute`; due to time limits, this functionality has not yet come to life.

We, as authors, strongly believe that implementing `PyRoute2WireGuard` for Darwin would resolve the remaining tests,
and as a result Vula could be fully functional on that system. We would love to finish this project in the next few
months, yet we believe that the scope of such integration should be taken on by a group of students as a semester project.

<a name="installation"></a>

## 1. Installation

Installation of Vula is relatively easy; for this scenario,
we propose installing it in two steps:

1. Clone the Vula repository—either your fork or the general `vula/vula` repository on codeberg.
2. Create a **venv** in which all necessary dependencies will be installed.

Speaking of which, you will need to install some system packages on your laptop:

```bash
brew install dbus pygobject3 wireguard-tools
```

Then install the Python dependencies directly into your freshly created venv with the next command:

```bash
pip install setuptools wheel build qrcode \
PyNaCl PyYAML Pillow hkdf highctidh pydbus \
click cryptography schema pyroute2 zeroconf \
rendez PyGObject pytest pytest-xdist
```

*Note: Keep track of the Python installation from Homebrew and the one your venv references; they
may target different versions and cause conflicts. For exmaple pytest is sometimes taken from the
Homebrew installation because it has higher priority in `$PATH`; in that case, use `python -m pytest`
to run the local one.*

### Failed Tests

**2022 situation:** The following tests failed on macOS.

```
FAILED vula/sys_pyroute2.py::vula.sys_pyroute2.Sys.get_stats
FAILED vula/wg.py::vula.wg._wg_interface_list
FAILED test/test_peer.py::test.test_peer.TestDescriptor_qrcode
FAILED test/test_sys_pyroute2.py::TestSys::test_start_stop_monitor - OSError:...
FAILED test/test_sys_pyroute2.py::TestSys::test_monitor_newneigh - OSError: [...
FAILED test/test_sys_pyroute2.py::TestSys::test_monitor_netlink_msg - OSError...
FAILED test/test_sys_pyroute2.py::TestSys::test_monitor_bug - OSError: [Errno...
FAILED test/test_verify.py::TestVerifyCommands::test_my_vk - AttributeError: ...
```

**2025 situation with `IPRoute` fix:**

```log
FAILED vula/sys_pyroute2.py::vula.sys_pyroute2.Sys.get_stats
FAILED vula/wg.py::vula.wg.Interface.__init__
FAILED vula/wg.py::vula.wg.Interface._get_link
FAILED test/test_wg.py::TestInterface::test_sync_interface - OSError: [Errno 47] Address family not supported by protocol family
FAILED test/test_wg.py::TestInterface::test_interface_query - OSError: [Errno 47] Address family not supported by protocol family
FAILED test/test_wg.py::TestInterface::test_peers_by_pubkey - OSError: [Errno 47] Address family not supported by protocol family
FAILED test/test_wg.py::TestInterface::test_apply_peerconfig - OSError: [Errno 47] Address family not supported by protocol family
```

All of these errors are related to the WireGuard class.

This result was achieved following the next set of commands on the arm64 machine:

```bash
brew install dbus pygobject3 wireguard-tools
git clone https://codeberg.org/vula/vula


python3 -m venv venv-vula
source venv-vula/bin/activate

pip install vula_libnss
pip install setuptools wheel build qrcode \
PyNaCl PyYAML Pillow hkdf highctidh pydbus \
click cryptography schema pyroute2 zeroconf \
rendez PyGObject pytest pytest-xdist

cd vula

pytest -v 
```

Versions at the time of the last test, 
including the Vula version defined by commit `6b36f4b578`.

The versions of Python and dependencies are as follows:
```bash
(venv-vula) bohdanpotuzhnyi@MBP-2 applicationSecurity % python --version
Python 3.13.4
(venv-vula) bohdanpotuzhnyi@MBP-2 applicationSecurity % pip list
Package            Version
------------------ --------------
asgiref            3.8.1
blinker            1.9.0
build              1.2.2.post1
certifi            2025.4.26
cffi               1.17.1
charset-normalizer 3.4.2
click              8.2.1
cryptography       45.0.4
execnet            2.1.1
Flask              3.1.1
glib               1.0.0
highctidh          1.0.2025051200
hkdf               0.0.3
idna               3.10
ifaddr             0.2.0
iniconfig          2.1.0
itsdangerous       2.2.0
Jinja2             3.1.6
MarkupSafe         3.0.2
packaging          25.0
pillow             11.2.1
pip                25.1.1
pluggy             1.6.0
pycairo            1.28.0
pycparser          2.22
pydbus             0.6.0
Pygments           2.19.1
PyGObject          3.52.3
pymonocypher       4.0.2.5
PyNaCl             1.5.0
pyproject_hooks    1.2.0
pyroute2           0.9.2
PySocks            1.7.1
pytest             8.4.0
pytest-xdist       3.7.0
PyYAML             6.0.2
qrcode             8.2
rendez             1.2.1
requests           2.32.4
schema             0.7.7
setuptools         80.9.0
stem               1.8.2
toml               0.10.2
urllib3            2.4.0
vula               0.2.2024011000
Werkzeug           3.1.3
wheel              0.45.1
zeroconf           0.147.0
```

<a name="debugging"></a>

## 3. Trying to Run in Monolithic Mode

It doesn't make sense to try to run Vula in the normal mode yet, but we can do it in **monolithic** mode.

Before running the next script, modify `__main__.py` to comment out the import of `dbus`, as only `pydbus` is
supported on macOS.

```bash
python -m vula.__main__ organize --keys-file "$HOME/.vula-organize-keys.yaml" run --monolithic
```

As we can see from the following logs, monolithic mode also expects a working `PyRoute2WireGuard()` for Darwin.

```
 self._wg = PyRoute2WireGuard()
 ...
 _socket.socket.__init__(self, family, type, proto, fileno)
    ~~~~~~~~~~~~~~~~~~~~~~~^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
OSError: [Errno 47] Address family not supported by protocol family
```

<a name="conclusion"></a>

## 4. Why the cross platform wireguard API does not work on MacOS?

### Apple's Network-Extension Security Model

macOS encloses all third-party VPN code inside the *Network Extension* framework, a security boundary that prohibits
direct interaction with low-level packet devices such as `/dev/net/tun`. Only binaries signed with Apple-issued
entitlements may create or control virtual interfaces, and those interfaces must be instantiated through the
framework's high-level APIs rather than raw `ioctl` calls. Consequently, the control paths assumed by portable
WireGuard tooling—“open device, configure it, push packets” — are unavailable on macOS unless the code is rewritten
to speak Network Extension and distributed with the requisite entitlements.

### Absence of an In-Kernel WireGuard Driver

On Linux, *Windows ≥ 10 2004*, and several BSDs, WireGuard ships as a kernel module that exposes its state
through system-specific control sockets (netlink, `wg` device files, or Windows Wintun IOCTLs). macOS, by
contrast, has no in-kernel implementation; the reference port (`wireguard-go`) runs entirely in user space
and shuttles packets through the generic *UTUN* interface provided by the operating system. Cross-platform
libraries such as **wgctrl** or **libwg**, which assume a kernel driver with a well-known control socket,
therefore find no endpoint to query or configure on macOS.

### Lack of a Uniform Management Socket

Linux exposes a stable, text-based configuration channel via `/var/run/wireguard/<ifname>.sock` (or the character
device `/dev/wireguard`), which portable APIs treat as their single source of truth. macOS offers no analogous, publicly
documented interface: configuration and runtime statistics live exclusively inside the Network Extension process,
accessible only through Apple's Objective-C/Swift APIs. Any “write once, run anywhere” WireGuard management layer
must therefore introduce macOS-specific glue code—or abandon the notion of a truly unified interface.

### Implications for Cross-Platform API Design

Taken together, these factors mean that a *single* WireGuard API cannot achieve functional parity across platforms
without conditional compilation and separate backend implementations:

| Platform | Packet Path                     | Control Path               | Privilege Boundary        |
| -------- | ------------------------------- | -------------------------- | ------------------------- |
| Linux    | Kernel module → `wg` netdevice  | Netlink / `/dev/wireguard` | `CAP_NET_ADMIN`           |
| Windows  | Kernel driver → Wintun          | IOCTL via DeviceIoControl  | Administrator / SERVICE   |
| macOS    | Userspace `wireguard-go` → UTUN | Network Extension API      | Apple-granted entitlement |

For library authors, the practical outcome is clear: **“cross-platform” WireGuard management requires
platform-specific back-ends, not merely portable build flags.**

## 5. Conclusion

If you follow all the instructions provided in this updated state of Vula on macOS, you will be able to run the tests
and identify the ones that fail due to missing `pyroute2` functionality. After that, you can also run Vula in monolithic
mode.

### Next Steps

As stated earlier, the only next step is to add Darwin support to `PyRoute2WireGuard`; in the authors' opinion, nothing
else can be done. After that, additional testing will likely be required to ensure that Vula works on the system, and
some extra tests might be introduced to verify its correctness on macOS.

### Considerations

It seems that a group of students has done a lot of work re-implementing Vula in Go; if they succeed, modifications to
the `pyroute2` library might no longer be needed, as it could work out of the box—and even on other systems such as
Windows, Android, and iOS.