package organize

import (
	"fmt"
	"strings"
	"testing"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"codeberg.org/vula/vula/contrib/vula-go/internal/schema"
	"gopkg.in/yaml.v3"
)

func TestPrefsManipulation(t *testing.T) {
	resultYaml := `pin_new_peers: false
auto_repair: true
subnets_allowed:
    - 10.0.0.0/8
    - 192.168.0.0/16
    - 172.16.0.0/12
    - 169.254.0.0/16
subnets_forbidden: []
iface_prefix_allowed:
    - en
    - eth
    - wl
    - thunderbolt
accept_nonlocal: false
local_domains:
    - local.
ephemeral_mode: false
accept_default_route: true
overwrite_unpinned: true
expire_time: 3600
primary_ip: null
record_events: false
enable_ipv6: false
enable_ipv4: false
`

	prefs := schema.Prefs{}

	err := setPref(&prefs, "pin_new_peers", "false")
	if err != nil {
		t.Errorf("error setting pref: %v", err)
	}

	err = setPref(&prefs, "auto_repair", "true")
	if err != nil {
		t.Errorf("error setting pref: %v", err)
	}

	err = addPref(&prefs, "subnets_allowed", "10.0.0.0/8")
	if err != nil {
		t.Errorf("error adding to pref: %v", err)
	}

	err = addPref(&prefs, "subnets_allowed", "192.168.0.0/16")
	if err != nil {
		t.Errorf("error adding to pref: %v", err)
	}

	err = addPref(&prefs, "subnets_allowed", "172.16.0.0/12")
	if err != nil {
		t.Errorf("error adding to pref: %v", err)
	}

	err = addPref(&prefs, "subnets_allowed", "169.254.0.0/16")
	if err != nil {
		t.Errorf("error adding to pref: %v", err)
	}

	err = addPref(&prefs, "iface_prefix_allowed", "en")
	if err != nil {
		t.Errorf("error adding to pref: %v", err)
	}

	err = addPref(&prefs, "iface_prefix_allowed", "eth")
	if err != nil {
		t.Errorf("error adding to pref: %v", err)
	}

	err = addPref(&prefs, "iface_prefix_allowed", "wl")
	if err != nil {
		t.Errorf("error adding to pref: %v", err)
	}

	err = addPref(&prefs, "iface_prefix_allowed", "thunderbolt")
	if err != nil {
		t.Errorf("error adding to pref: %v", err)
	}

	err = setPref(&prefs, "accept_nonlocal", "false")
	if err != nil {
		t.Errorf("error setting pref: %v", err)
	}

	err = addPref(&prefs, "local_domains", "local.")
	if err != nil {
		t.Errorf("error adding to pref: %v", err)
	}

	err = setPref(&prefs, "ephemeral_mode", "false")
	if err != nil {
		t.Errorf("error setting pref: %v", err)
	}

	err = setPref(&prefs, "accept_default_route", "true")
	if err != nil {
		t.Errorf("error setting pref: %v", err)
	}

	err = setPref(&prefs, "overwrite_unpinned", "true")
	if err != nil {
		t.Errorf("error setting pref: %v", err)
	}

	err = setPref(&prefs, "expire_time", "3600")
	if err != nil {
		t.Errorf("error setting pref: %v", err)
	}

	err = setPref(&prefs, "record_events", "false")
	if err != nil {
		t.Errorf("error setting pref: %v", err)
	}

	serialized, err := yaml.Marshal(prefs)
	if err != nil {
		t.Errorf("Error serializing prefs: %v", err)
	}

	if string(serialized) != resultYaml {
		t.Errorf("serialized = %s", string(serialized))
	}
}

func TestPrefsAddForbiddenAllowedNet(t *testing.T) {
	prefs := schema.Prefs{}

	err := addPref(&prefs, "subnets_forbidden", "10.0.0.0/8")
	if err != nil {
		t.Errorf("error adding to list pref: %v", err)
	}

	err = addPref(&prefs, "subnets_allowed", "10.0.0.0/8")
	if err == nil {
		t.Errorf("net added to allow list despite being on forbidden list")
	}
}

func TestPrefsAddAllowedForbiddenNet(t *testing.T) {
	prefs := schema.Prefs{}

	err := addPref(&prefs, "subnets_allowed", "192.168.0.0/16")
	if err != nil {
		t.Errorf("error adding to list pref: %v", err)
	}

	err = addPref(&prefs, "subnets_forbidden", "192.168.0.0/16")
	if err == nil {
		t.Errorf("net added to forbidden list despite being on allow list")
	}
}

func TestPrefsAddNetTwice(t *testing.T) {
	prefs := schema.Prefs{}

	err := addPref(&prefs, "subnets_allowed", "10.0.0.0/24")
	if err != nil {
		t.Errorf("error adding to list pref: %v", err)
	}

	err = addPref(&prefs, "subnets_allowed", "10.0.0.10/24")
	if err != nil {
		t.Errorf("error adding to list pref: %v", err)
	}
}

func TestPrefsAddInvalidIfacePrefix(t *testing.T) {
	prefs := schema.Prefs{}

	err := addPref(&prefs, "iface_prefix_allowed", "my%prefix")
	if err == nil {
		t.Errorf("added invalid iface prefix")
	}
}

func TestPrefsAddInvalidLocalDomain(t *testing.T) {
	prefs := schema.Prefs{}

	err := addPref(&prefs, "local_domains", "my%local.domain")
	if err == nil {
		t.Errorf("added invalid local domain")
	}
}

func TestPrefsSetUnknownPref(t *testing.T) {
	prefs := schema.Prefs{}

	err := addPref(&prefs, "some_other_pref", "foobar")
	if err == nil {
		t.Errorf("set unknown pref")
	}
}

func setPref(p *schema.Prefs, key string, value string) error {
	err := schema.Write(value, []string{key}, p)
	if err != nil {
		return err
	}
	c := schema.ImportPrefs(p)
	return verifyPrefs(&c)
}

func addPref(p *schema.Prefs, key string, value any) error {
	err := schema.Add(value, []string{key}, p)
	if err != nil {
		return err
	}
	c := schema.ImportPrefs(p)
	return verifyPrefs(&c)
}

func verifyPrefs(p *core.Prefs) (err error) {
	messages := validatePrefs(p)
	if len(messages) > 0 {
		err = fmt.Errorf("invalid prefs: %s", strings.Join(messages, ", "))
	}
	return
}
