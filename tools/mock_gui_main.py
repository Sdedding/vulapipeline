#!/usr/bin/env python3
"""
A simple mock main that launches the GUI with no real functionality.
It uses a fake DataProvider so that DBus/system calls are bypassed.
"""

# -----------------------------------------------------
# 1. A mock/fake DataProvider that returns dummy data
# -----------------------------------------------------
class MockDataProvider:
    def get_peers(self):
        # Return an empty list or minimal fake peers:
        return []

    def get_prefs(self):
        # Return empty or fake preferences
        return {
            "pin_new_peers": False,
            "auto_repair": True,
            "subnets_allowed": ["192.168.0.0/24"],
            "subnets_forbidden": [],
            "iface_prefix_allowed": ["eth", "wl"],
            "accept_nonlocal": False,
            "local_domains": ["local."],
            "ephemeral_mode": False,
            "accept_default_route": True,
            "record_events": False,
            "expire_time": 3600,
            "overwrite_unpinned": True,
        }

    def get_status(self):
        # Pretend all services are "inactive" or "active" to see the UI
        return {
            "publish": "inactive",
            "discover": "inactive",
            "organize": "inactive",
        }

    def our_latest_descriptors(self):
        # Return an empty JSON object or minimal structure
        return "{}"

    # The “write” type methods can just be no-ops:
    def delete_peer(self, peer_vk):
        pass

    def rename_peer(self, peer_vk, name):
        pass

    def pin_and_verify(self, peer_vk, peer_name):
        pass

    def add_peer(self, peer_vk, ip):
        pass

    def set_pref(self, pref, value):
        return ""

    def add_pref(self, pref, value):
        return ""

    def remove_pref(self, pref, value):
        return ""


# -----------------------------------------------------
# 2. Patch or override the actual DataProvider in code
# -----------------------------------------------------
def patch_dataprovider():
    """Replace the real DataProvider with our mock so that GUI calls the mock methods."""
    import vula.frontend.dataprovider as dp
    import vula.frontend

    dp.DataProvider = MockDataProvider
    vula.frontend.DataProvider = MockDataProvider

# -----------------------------------------------------
# 3. Patch or override the actual constants in code
# -----------------------------------------------------  
def patch_constants():
    import vula.frontend.constants as c
    c.IMAGE_BASE_PATH = "/workspaces/vula/misc/images/"

# -----------------------------------------------------
# 4. Run the GUI code directly
# -----------------------------------------------------
def main():
    # 3.1 patch the data provider
    patch_dataprovider()
    
    # 3.2 patch the constants
    patch_constants()

    # 3.3 import the GUI “main” after patching
    from vula.frontend.ui import App

    # 3.4 create and launch the Tkinter application
    root = App()
    root.title("Vula-GUI (Mocked)")
    root.mainloop()


if __name__ == "__main__":
    main()
