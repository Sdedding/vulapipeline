package organize

import (
	"testing"

	"codeberg.org/vula/vula/contrib/vula-go/internal/schema"
)

const (
	testDescriptor         = "c=cBVKup6b9dM6hfY0pE81fCKPJ6EFVvT7m+Gkt/W7gIHhBl50fdKZzT5feHACzJXDRzhxYicoyi358tREqhcyWw==; dt=86400; e=0; hostname=vula-bookworm-test2.local.; pk=6T2K6Xcmlsr1XQVZTAHrZs/d9v3IadKYI+74559/3Aw=; port=5354; r=; s=PuDfyhWpftSbWUMMydt1Qv7o618KIli9ncxUkcPP8yqaspDXa0jJUnwNwydEpXjVfY96BmVu5Jwba8ahZPzBDA==; v4a=10.89.0.3; v6a=fdff:ffff:ffdf:e436:dfba:4f29:bcbf:6af8,fe80::cc69:7dff:fe6b:9e79,fd54:f27a:17c1:3a61::3; vf=1743985213; vk=Gy+arU0cowJC2vek9EnoGHVSQxUl5Qv1LUrDL/WjGos=;"
	testDescriptorUnsigned = "c=NnoGEZ4W+d6TE22+Qyau0LF513FM43EagOP9aiSX9KhTCS1Gryt7qDoM04j7p0KQRJxwkcPEO/MpIJE5/bJKYQ==; dt=86400; e=0; hostname=vula-bookworm-test1.local.; p=fdff::1; pk=3w5/xje5jsdUCX30JfS/L/bMuwZRniK69dAVprN7t3c=; port=5354; r=; v4a=10.89.0.2; v6a=fdff:ffff:ffdf:989f:24cf:bda:1262:cfc6,fe80::bc92:4dff:fe82:30d,fd54:f27a:17c1:3a61::2; vf=1743974365; vk=afToKyN29ubu4DkhUMLoGIt5WjbsgEHYuccNtxvbjmA=;"
)

func TestVerifyDescriptor(t *testing.T) {
	// arrange
	d, err := schema.ParseDescriptorString(testDescriptor)
	if err != nil {
		t.Fatal(err)
	}

	// act
	ok := verifyDescriptorSignature(&d)

	// assert
	if !ok {
		t.Errorf("descriptor signature invalid")
	}
}

func TestSignDescriptor(t *testing.T) {
	// arrange
	keys, err := genKeys()
	if err != nil {
		t.Fatal(err)
	}
	d, err := schema.ParseDescriptorString(testDescriptor)
	if err != nil {
		t.Fatal(err)
	}
	d.VK = keys.VkEd25519.PK

	// act
	err = signDescriptor(&d, keys.VkEd25519.SK)
	if err != nil {
		t.Error(err)
	}

	// assert
	if !verifyDescriptorSignature(&d) {
		t.Errorf("invalid signature")
	}
}
