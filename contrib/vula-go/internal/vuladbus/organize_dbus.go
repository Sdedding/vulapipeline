package vuladbus

import (
	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"codeberg.org/vula/vula/contrib/vula-go/internal/util"
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
)

const (
	organizeSyncName              = "local.vula.organize1.Sync"
	organizeDebugName             = "local.vula.organize1.Debug"
	organizeProcessDescriptorName = "local.vula.organize1.ProcessDescriptor"
	organizePrefsName             = "local.vula.organize1.Prefs"
	organizePeersName             = "local.vula.organize1.Peers"
)

var organizeIntrospectable = introspect.NewIntrospectable(&introspect.Node{
	Interfaces: []introspect.Interface{
		introspect.IntrospectData,
		{
			Name: organizeSyncName,
			Methods: []introspect.Method{
				{
					Name: "sync",
					Args: []introspect.Arg{
						{Name: "dry_run", Type: "b", Direction: "in"},
						{Name: "response", Type: "as", Direction: "out"},
					},
				},
			},
		},
		{
			Name: organizeDebugName,
			Methods: []introspect.Method{
				{
					Name: "dump_state",
					Args: []introspect.Arg{
						{Name: "interactive", Type: "b", Direction: "in"},
						{Name: "response", Type: "s", Direction: "out"},
					},
				},
				{
					Name: "test_auth",
					Args: []introspect.Arg{
						{Name: "interactive", Type: "b", Direction: "in"},
						{Name: "response", Type: "s", Direction: "out"},
					},
				},
			},
		},
		{
			Name: organizePeersName,
			Methods: []introspect.Method{
				{
					Name: "show_peer",
					Args: []introspect.Arg{
						{Name: "query", Type: "s", Direction: "in"},
						{Name: "response", Type: "s", Direction: "out"},
					},
				},
				{
					Name: "peer_descriptor",
					Args: []introspect.Arg{
						{Name: "query", Type: "s", Direction: "in"},
						{Name: "response", Type: "s", Direction: "out"},
					},
				},
				{
					Name: "peer_ids",
					Args: []introspect.Arg{
						{Name: "which", Type: "s", Direction: "in"},
						{Name: "response", Type: "as", Direction: "out"},
					},
				},
				{
					Name: "rediscover",
					Args: []introspect.Arg{
						{Name: "response", Type: "s", Direction: "out"},
					},
				},
				{
					Name: "set_peer",
					Args: []introspect.Arg{
						{Name: "vk", Type: "s", Direction: "in"},
						{Name: "path", Type: "as", Direction: "in"},
						{Name: "value", Type: "s", Direction: "in"},
						{Name: "response", Type: "s", Direction: "out"},
					},
				},
				{
					Name: "remove_peer",
					Args: []introspect.Arg{
						{Name: "vk", Type: "s", Direction: "in"},
						{Name: "response", Type: "s", Direction: "out"},
					},
				},
				{
					Name: "peer_addr_add",
					Args: []introspect.Arg{
						{Name: "vk", Type: "s", Direction: "in"},
						{Name: "value", Type: "s", Direction: "in"},
						{Name: "response", Type: "s", Direction: "out"},
					},
				},
				{
					Name: "peer_addr_del",
					Args: []introspect.Arg{
						{Name: "vk", Type: "s", Direction: "in"},
						{Name: "value", Type: "s", Direction: "in"},
						{Name: "response", Type: "s", Direction: "out"},
					},
				},
				{
					Name: "our_latest_descriptors",
					Args: []introspect.Arg{
						{Name: "descriptors", Type: "s", Direction: "out"},
					},
				},
				{
					Name: "get_vk_by_name",
					Args: []introspect.Arg{
						{Name: "hostname", Type: "s", Direction: "in"},
						{Name: "response", Type: "s", Direction: "out"},
					},
				},
				{
					Name: "verify_and_pin_peer",
					Args: []introspect.Arg{
						{Name: "vk", Type: "s", Direction: "in"},
						{Name: "hostname", Type: "s", Direction: "in"},
						{Name: "response", Type: "s", Direction: "out"},
					},
				},
			},
		},
		{
			Name: organizeProcessDescriptorName,
			Methods: []introspect.Method{
				{
					Name: "process_descriptor_string",
					Args: []introspect.Arg{
						{Name: "descriptor", Type: "s", Direction: "in"},
						{Name: "response", Type: "s", Direction: "out"},
					},
				},
			},
		},
		{
			Name: organizePrefsName,
			Methods: []introspect.Method{
				{
					Name: "show_prefs",
					Args: []introspect.Arg{
						{Name: "response", Type: "s", Direction: "out"},
					},
				},
				{
					Name: "set_pref",
					Args: []introspect.Arg{
						{Name: "response", Type: "s", Direction: "out"},
						{Name: "pref", Type: "s", Direction: "in"},
						{Name: "value", Type: "s", Direction: "in"},
					},
				},
				{
					Name: "add_pref",
					Args: []introspect.Arg{
						{Name: "response", Type: "s", Direction: "out"},
						{Name: "pref", Type: "s", Direction: "in"},
						{Name: "value", Type: "s", Direction: "in"},
					},
				},
				{
					Name: "remove_pref",
					Args: []introspect.Arg{
						{Name: "response", Type: "s", Direction: "out"},
						{Name: "pref", Type: "s", Direction: "in"},
						{Name: "value", Type: "s", Direction: "in"},
					},
				},
				{
					Name: "release_gateway",
					Args: []introspect.Arg{
						{Name: "response", Type: "s", Direction: "out"},
					},
				},
			},
		},
	},
})

type organizeSync struct {
	o core.OrganizeSync
}

func (d *organizeSync) Sync(dryRun bool) ([]string, *dbus.Error) {
	core.LogDebugf("Dbus organize call: Sync(dryRun=%v)", dryRun)
	response, err := d.o.Sync(dryRun)
	if err != nil {
		return response, dbus.MakeFailedError(err)
	}
	return response, nil
}

type organizeDebug struct {
	o core.OrganizeDebug
}

func (d *organizeDebug) DumpState(interactive bool) (string, *dbus.Error) {
	core.LogDebugf("Dbus organize call: DumpState(interactive=%v)", interactive)
	response, err := d.o.DumpState(interactive)
	if err != nil {
		return response, dbus.MakeFailedError(err)
	}
	return response, nil
}

func (d *organizeDebug) TestAuth(interactive bool) (string, *dbus.Error) {
	core.LogDebugf("Dbus organize call: TestAuth(interactive=%v)", interactive)
	response, err := d.o.TestAuth(interactive)
	if err != nil {
		return response, dbus.MakeFailedError(err)
	}
	return response, nil
}

type organizePeers struct {
	o core.OrganizePeers
}

func (d *organizePeers) ShowPeer(query string) (string, *dbus.Error) {
	core.LogDebugf("Dbus organize call: ShowPeer(query=%q)", query)
	response, err := d.o.ShowPeer(query)
	if err != nil {
		return response, dbus.MakeFailedError(err)
	}
	return response, nil
}

func (d *organizePeers) PeerDescriptor(query string) (string, *dbus.Error) {
	core.LogDebugf("Dbus organize call: PeerDescriptor(query=%q)", query)
	response, err := d.o.PeerDescriptor(query)
	if err != nil {
		return response, dbus.MakeFailedError(err)
	}
	return response, nil
}

func (d *organizePeers) PeerIds(which string) ([]string, *dbus.Error) {
	core.LogDebugf("Dbus organize call: PeerIds(which=%q)", which)
	response, err := d.o.PeerIds(which)
	if err != nil {
		return response, dbus.MakeFailedError(err)
	}
	return response, nil
}

func (d *organizePeers) Rediscover() (string, *dbus.Error) {
	core.LogDebugf("Dbus organize call: Rediscover()")
	response, err := d.o.Rediscover()
	if err != nil {
		return response, dbus.MakeFailedError(err)
	}
	return response, nil
}

func (d *organizePeers) SetPeer(vk string, path []string, value string) (string, *dbus.Error) {
	core.LogDebugf("Dbus organize call: SetPeer(vk=%q, path=%v, value=%q)", vk, path, value)
	response, err := d.o.SetPeer(vk, path, value)
	if err != nil {
		return response, dbus.MakeFailedError(err)
	}
	return response, nil
}

func (d *organizePeers) RemovePeer(vk string) (string, *dbus.Error) {
	core.LogDebugf("Dbus organize call: RemovePeer(vk=%q)", vk)
	response, err := d.o.RemovePeer(vk)
	if err != nil {
		return response, dbus.MakeFailedError(err)
	}
	return response, nil
}

func (d *organizePeers) PeerAddrAdd(vk string, value string) (string, *dbus.Error) {
	core.LogDebugf("Dbus organize call: PeerAddrAdd(vk=%q, value=%q)", vk, value)
	response, err := d.o.PeerAddrAdd(vk, value)
	if err != nil {
		return response, dbus.MakeFailedError(err)
	}
	return response, nil
}

func (d *organizePeers) PeerAddrDel(vk string, value string) (string, *dbus.Error) {
	core.LogDebugf("Dbus organize call: PeerAddrDel(vk=%q, value=%q)", vk, value)
	response, err := d.o.PeerAddrDel(vk, value)
	if err != nil {
		return response, dbus.MakeFailedError(err)
	}
	return response, nil
}

func (d *organizePeers) OurLatestDescriptors() (string, *dbus.Error) {
	core.LogDebugf("Dbus organize call: OurLatestDescriptors()")
	response, err := d.o.OurLatestDescriptors()
	if err != nil {
		return response, dbus.MakeFailedError(err)
	}
	return response, nil
}

func (d *organizePeers) GetVkByName(hostname string) (string, *dbus.Error) {
	core.LogDebugf("Dbus organize call: GetVkByName(hostname=%q)", hostname)
	response, err := d.o.GetVkByName(hostname)
	if err != nil {
		return response, dbus.MakeFailedError(err)
	}
	return response, nil
}

func (d *organizePeers) VerifyAndPinPeer(vk string, hostname string) (string, *dbus.Error) {
	core.LogDebugf("Dbus organize call: VerifyAndPinPeer(vk=%q, hostname=%q)", vk, hostname)
	response, err := d.o.VerifyAndPinPeer(vk, hostname)
	if err != nil {
		return response, dbus.MakeFailedError(err)
	}
	return response, nil
}

type organizeProcessDescriptor struct {
	o core.OrganizeProcessDescriptor
}

func (d organizeProcessDescriptor) ProcessDescriptorString(descriptor string) (string, *dbus.Error) {
	core.LogDebugf("Dbus organize call: ProcessDescriptorString(descriptor=%q)", descriptor)
	response, err := d.o.ProcessDescriptorString(descriptor)
	if err != nil {
		core.LogWarn("Failed to process descriptor:", err)
		return response, dbus.MakeFailedError(err)
	}
	return response, nil
}

type organizePrefs struct {
	o core.OrganizePrefs
}

func (d *organizePrefs) ShowPrefs() (string, *dbus.Error) {
	core.LogDebugf("Dbus organize call: ShowPrefs()")
	response, err := d.o.ShowPrefs()
	if err != nil {
		return response, dbus.MakeFailedError(err)
	}
	return response, nil
}

func (d *organizePrefs) SetPref(prefKey string, value string) (string, *dbus.Error) {
	core.LogDebugf("Dbus organize call: SetPref(prefKey=%q, value=%q)", prefKey, value)
	response, err := d.o.SetPref(prefKey, value)
	if err != nil {
		return response, dbus.MakeFailedError(err)
	}
	return response, nil
}

func (d *organizePrefs) AddPref(prefKey string, value string) (string, *dbus.Error) {
	core.LogDebugf("Dbus organize call: AddPref(prefKey=%q, value=%q)", prefKey, value)
	response, err := d.o.AddPref(prefKey, value)
	if err != nil {
		return response, dbus.MakeFailedError(err)
	}
	return response, nil
}

func (d *organizePrefs) RemovePref(prefKey string, value string) (string, *dbus.Error) {
	core.LogDebugf("Dbus organize call: RemovePref(prefKey=%q, value=%q)", prefKey, value)
	response, err := d.o.RemovePref(prefKey, value)
	if err != nil {
		return response, dbus.MakeFailedError(err)
	}
	return response, nil
}

func (d *organizePrefs) ReleaseGateway() (string, *dbus.Error) {
	core.LogDebugf("Dbus organize call: ReleaseGateway()")
	response, err := d.o.ReleaseGateway()
	if err != nil {
		return response, dbus.MakeFailedError(err)
	}
	return response, nil
}

func RunOrganizeServer(done <-chan struct{}, o core.Organize) error {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return err
	}
	defer util.CloseLog(conn)

	err = exportDbusOrganize(conn, o)
	if err != nil {
		return err
	}

	<-done
	return nil
}

type OrganizeClient struct {
	conn           *dbus.Conn
	organizeObject dbus.BusObject
}

func NewOrganizeClient() (*OrganizeClient, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}
	organizeObject := conn.Object(
		"local.vula.organize",
		"/local/vula/organize",
	)
	return &OrganizeClient{
		conn:           conn,
		organizeObject: organizeObject,
	}, nil
}

func (d *OrganizeClient) Close() error {
	return d.conn.Close()
}

func (d *OrganizeClient) DumpState(interactive bool) (string, error) {
	call := d.organizeObject.Call(organizeDebugName+".dump_state", 0, interactive)
	if call.Err != nil {
		return "", call.Err
	}
	var result string
	err := call.Store(&result)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (d *OrganizeClient) TestAuth(interactive bool) (string, error) {
	call := d.organizeObject.Call(organizeDebugName+".test_auth", 0, interactive)
	if call.Err != nil {
		return "", call.Err
	}
	var result string
	err := call.Store(&result)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (d *OrganizeClient) Sync(dryRun bool) ([]string, error) {
	call := d.organizeObject.Call(organizeSyncName+".sync", 0, dryRun)
	if call.Err != nil {
		return nil, call.Err
	}
	var result []string
	err := call.Store(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (d *OrganizeClient) ShowPeer(query string) (string, error) {
	call := d.organizeObject.Call(organizePeersName+".show_peer", 0, query)
	if call.Err != nil {
		return "", call.Err
	}
	var result string
	err := call.Store(&result)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (d *OrganizeClient) PeerDescriptor(query string) (string, error) {
	call := d.organizeObject.Call(organizePeersName+".peer_descriptor", 0, query)
	if call.Err != nil {
		return "", call.Err
	}
	var result string
	err := call.Store(&result)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (d *OrganizeClient) PeerIds(which string) ([]string, error) {
	call := d.organizeObject.Call(organizePeersName+".peer_ids", 0, which)
	if call.Err != nil {
		return nil, call.Err
	}
	var result []string
	err := call.Store(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (d *OrganizeClient) Rediscover() (string, error) {
	call := d.organizeObject.Call(organizePeersName+".rediscover", 0)
	if call.Err != nil {
		return "", call.Err
	}
	var result string
	err := call.Store(&result)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (d *OrganizeClient) SetPeer(vk string, path []string, value string) (string, error) {
	call := d.organizeObject.Call(organizePeersName+".set_peer", 0, vk, path, value)
	if call.Err != nil {
		return "", call.Err
	}
	var result string
	err := call.Store(&result)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (d *OrganizeClient) RemovePeer(vk string) (string, error) {
	call := d.organizeObject.Call(organizePeersName+".remove_peer", 0, vk)
	if call.Err != nil {
		return "", call.Err
	}
	var result string
	err := call.Store(&result)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (d *OrganizeClient) PeerAddrAdd(vk string, value string) (string, error) {
	call := d.organizeObject.Call(organizePeersName+".peer_addr_add", 0, vk, value)
	if call.Err != nil {
		return "", call.Err
	}
	var result string
	err := call.Store(&result)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (d *OrganizeClient) PeerAddrDel(vk string, value string) (string, error) {
	call := d.organizeObject.Call(organizePeersName+".peer_addr_del", 0, vk, value)
	if call.Err != nil {
		return "", call.Err
	}
	var result string
	err := call.Store(&result)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (d *OrganizeClient) OurLatestDescriptors() (string, error) {
	call := d.organizeObject.Call(organizePeersName+".our_latest_descriptors", 0)
	if call.Err != nil {
		return "", call.Err
	}
	var result string
	err := call.Store(&result)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (d *OrganizeClient) GetVkByName(hostname string) (string, error) {
	call := d.organizeObject.Call(organizePeersName+".get_vk_by_name", 0, hostname)
	if call.Err != nil {
		return "", call.Err
	}
	var result string
	err := call.Store(&result)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (d *OrganizeClient) VerifyAndPinPeer(vk string, hostname string) (string, error) {
	call := d.organizeObject.Call(organizePeersName+".verify_and_pin_peer", 0, vk, hostname)
	if call.Err != nil {
		return "", call.Err
	}
	var result string
	err := call.Store(&result)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (d *OrganizeClient) ShowPrefs() (string, error) {
	call := d.organizeObject.Call(organizePrefsName+".show_prefs", 0)
	if call.Err != nil {
		return "", call.Err
	}
	var result string
	err := call.Store(&result)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (d *OrganizeClient) SetPref(pref string, value string) (string, error) {
	call := d.organizeObject.Call(organizePrefsName+".set_pref", 0, pref, value)
	if call.Err != nil {
		return "", call.Err
	}
	var result string
	err := call.Store(&result)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (d *OrganizeClient) AddPref(pref string, value string) (string, error) {
	call := d.organizeObject.Call(organizePrefsName+".add_pref", 0, pref, value)
	if call.Err != nil {
		return "", call.Err
	}
	var result string
	err := call.Store(&result)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (d *OrganizeClient) RemovePref(pref string, value string) (string, error) {
	call := d.organizeObject.Call(organizePrefsName+".remove_pref", 0, pref, value)
	if call.Err != nil {
		return "", call.Err
	}
	var result string
	err := call.Store(&result)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (d *OrganizeClient) ReleaseGateway() (string, error) {
	call := d.organizeObject.Call(organizePrefsName+".release_gateway", 0)
	if call.Err != nil {
		return "", call.Err
	}
	var result string
	err := call.Store(&result)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (d *OrganizeClient) ProcessDescriptorString(descriptorString string) (string, error) {
	call := d.organizeObject.Call(organizeProcessDescriptorName+".process_descriptor_string", 0, descriptorString)
	if call.Err != nil {
		return "", call.Err
	}
	var result string
	err := call.Store(&result)
	return result, err
}

var _ core.OrganizeSync = &OrganizeClient{}
var _ core.OrganizePeers = &OrganizeClient{}
var _ core.OrganizeDebug = &OrganizeClient{}
var _ core.OrganizePrefs = &OrganizeClient{}
var _ core.OrganizeProcessDescriptor = &OrganizeClient{}

func exportDbusOrganize(conn *dbus.Conn, o core.Organize) (err error) {
	type export struct {
		v       any
		mapping map[string]string
		name    string
	}

	var exports = []export{
		{
			v:       &organizeSync{o},
			mapping: map[string]string{"Sync": "sync"},
			name:    organizeSyncName},
		{
			v:       &organizeDebug{o},
			mapping: map[string]string{"DumpState": "dump_state", "TestAuth": "test_auth"},
			name:    organizeDebugName},
		{
			v: &organizePeers{o},
			mapping: map[string]string{
				"ShowPeer":             "show_peer",
				"PeerDescriptor":       "peer_descriptor",
				"PeerIds":              "peer_ids",
				"Rediscover":           "rediscover",
				"SetPeer":              "set_peer",
				"RemovePeer":           "remove_peer",
				"PeerAddrAdd":          "peer_addr_add",
				"PeerAddrDel":          "peer_addr_del",
				"OurLatestDescriptors": "our_latest_descriptors",
				"GetVkByName":          "get_vk_by_name",
				"VerifyAndPinPeer":     "verify_and_pin_peer",
			},
			name: organizePeersName,
		},
		{
			v:       &organizeProcessDescriptor{o},
			mapping: map[string]string{"ProcessDescriptorString": "process_descriptor_string"},
			name:    organizeProcessDescriptorName,
		},
		{
			v: &organizePrefs{o},
			mapping: map[string]string{
				"ShowPrefs":      "show_prefs",
				"SetPref":        "set_pref",
				"AddPref":        "add_pref",
				"RemovePref":     "remove_pref",
				"ReleaseGateway": "release_gateway",
			},
			name: organizePrefsName,
		},
	}

	for i := range exports {
		export := &exports[i]
		err = conn.ExportWithMap(export.v, export.mapping, organizeDbusPath, export.name)
		if err != nil {
			return err
		}
	}

	err = conn.Export(organizeIntrospectable, organizeDbusPath, dbusIntrospectableName)
	if err != nil {
		return
	}
	err = acquireDbusName(conn, organizeDbusName)
	return
}
