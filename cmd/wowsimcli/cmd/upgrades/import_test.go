package upgrades

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

func encodeIndividualLink(settings *proto.IndividualSimSettings) (string, error) {
	data, err := googleProto.Marshal(settings)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return "https://wowsims.com/tbc/mage/#" + encoded, nil
}

func fixtureIndividualSettings() *proto.IndividualSimSettings {
	return &proto.IndividualSimSettings{
		ApiVersion: core.GetCurrentProtoVersion(),
		Settings: &proto.SimSettings{
			Iterations:   1000,
			Phase:        2,
			FixedRngSeed: 42,
		},
		RaidBuffs: &proto.RaidBuffs{
			ArcaneBrilliance: true,
			DivineSpirit:     proto.TristateEffect_TristateEffectRegular,
			Bloodlust:        true,
		},
		PartyBuffs: &proto.PartyBuffs{
			DraeneiRacialCaster: true,
		},
		Debuffs: &proto.Debuffs{
			JudgementOfWisdom: true,
			Misery:            true,
		},
		Tanks: []*proto.UnitReference{
			{Type: proto.UnitReference_Player, Index: 0},
		},
		Encounter: &proto.Encounter{
			Duration: 180,
			Targets: []*proto.Target{
				{
					Id:    1,
					Name:  "Boss",
					Level: 73,
				},
			},
		},
		Player: &proto.Player{
			ApiVersion:  core.GetCurrentProtoVersion(),
			Name:        "TestMage",
			Class:       proto.Class_ClassMage,
			Race:        proto.Race_RaceTroll,
			Profession1: proto.Profession_Tailoring,
			Profession2: proto.Profession_Enchanting,
			TalentsString: "",
			Rotation: &proto.APLRotation{
				Type: proto.APLRotation_TypeAPL,
				PriorityList: []*proto.APLListItem{
					{
						Action: &proto.APLAction{
							Action: &proto.APLAction_CastSpell{
								CastSpell: &proto.APLActionCastSpell{
									SpellId: &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: 27074}}, // Scorch
								},
							},
						},
					},
				},
			},
			Spec: &proto.Player_Mage{
				Mage: &proto.Mage{
					Options: &proto.Mage_Options{
						ClassOptions: &proto.MageOptions{
							DefaultMageArmor: proto.MageArmor_MageArmorMageArmor,
						},
					},
				},
			},
			Equipment: &proto.EquipmentSpec{
				Items: []*proto.ItemSpec{
					{Id: 24266}, // Head: Spellstrike Hood
					{Id: 28530}, // Neck: Brooch of Unquenchable Fury
					{Id: 29079}, // Shoulder: Pauldrons of the Aldor
					{Id: 28766}, // Back: Ruby Drape of the Mystic
					{Id: 29077}, // Chest: Robes of the Aldor
					{Id: 24250}, // Wrist: Bracers of Havok
					{Id: 29078}, // Hands: Gloves of the Aldor
					{Id: 30038}, // Waist: Belt of Blasting
					{Id: 29080}, // Legs: Leggings of the Aldor
					{Id: 30037}, // Feet: Boots of Blasting
					{Id: 30109}, // Finger1: Ring of Endless Coils
					{Id: 29305}, // Finger2: Band of the Eternal Sage
					{Id: 28785}, // Trinket1: The Lightning Capacitor
					{Id: 29370}, // Trinket2: Icon of the Silver Crescent
					{Id: 28770}, // MainHand: Nathrezim Mindblade
					{Id: 29272}, // OffHand: Talisman of Kalecgos
					{Id: 28673}, // Ranged: Tirisfal Wand of Ascendancy
				},
			},
		},
	}
}

func initFixtures(t *testing.T) {
	if t != nil {
		t.Helper()
	}
	dir := filepath.Join("testdata")
	if err := os.MkdirAll(dir, 0755); err != nil {
		if t != nil {
			t.Fatal(err)
		}
		panic(err)
	}
	linkPath := filepath.Join(dir, "fixed_individual_link.txt")
	summaryPath := filepath.Join(dir, "fixed_import_summary.json")

	if _, err := os.Stat(linkPath); os.IsNotExist(err) {
		settings := fixtureIndividualSettings()
		link, err := encodeIndividualLink(settings)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(linkPath, []byte(link), 0644); err != nil {
			t.Fatal(err)
		}

		imported, err := Import(link)
		if err != nil {
			t.Fatal(err)
		}

		summaryJSON, err := json.MarshalIndent(imported.Character, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(summaryPath, summaryJSON, 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func readFixture(t *testing.T, filename string) string {
	if t != nil {
		t.Helper()
	}
	data, err := os.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", filename, err)
	}
	return string(bytes.TrimSpace(data))
}

func mustImportFixture(t *testing.T) *ImportedSettings {
	if t != nil {
		t.Helper()
	}
	link := readFixture(t, "fixed_individual_link.txt")
	imported, err := Import(link)
	if err != nil {
		if t != nil {
			t.Fatalf("Import failed: %v", err)
		}
		panic(err)
	}
	return imported
}

func mustMarshal(t *testing.T, m googleProto.Message) []byte {
	t.Helper()
	data, err := googleProto.MarshalOptions{Deterministic: true}.Marshal(m)
	if err != nil {
		t.Fatalf("failed to marshal message: %v", err)
	}
	return data
}

func TestImportDecodesFixedIndividualLink(t *testing.T) {
	initFixtures(t)
	link := readFixture(t, "fixed_individual_link.txt")
	imported, err := Import(link)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := imported.Character.Class, proto.Class_ClassMage.String(); got != want {
		t.Fatalf("class = %q, want %q", got, want)
	}
	if got, want := imported.Character.Spec, "Mage"; got != want {
		t.Fatalf("spec = %q, want %q", got, want)
	}
	if got, want := imported.Character.EquippedItems, 17; got != want {
		t.Fatalf("equipped items = %d, want %d", got, want)
	}
	if got, want := imported.Character.Phase, int32(2); got != want {
		t.Fatalf("phase = %d, want %d", got, want)
	}
	if got, want := imported.SimulatorVersion, SimulatorRevision; got != want {
		t.Fatalf("simulator revision = %q, want %q", got, want)
	}
	if got, want := imported.DatabaseVersion, DatabaseRevision; got != want {
		t.Fatalf("database revision = %q, want %q", got, want)
	}
	if imported.SettingsDigest == "" {
		t.Fatal("settings digest is empty")
	}

	// Verify against fixed_import_summary.json
	summaryRaw := readFixture(t, "fixed_import_summary.json")
	var expectedSummary CharacterSummary
	if err := json.Unmarshal([]byte(summaryRaw), &expectedSummary); err != nil {
		t.Fatalf("failed to unmarshal summary fixture: %v", err)
	}
	if imported.Character.Name != expectedSummary.Name || imported.Character.Class != expectedSummary.Class {
		t.Fatalf("summary mismatch: got %+v, want %+v", imported.Character, expectedSummary)
	}
}

func TestImportAcceptsExportWithoutSimSettings(t *testing.T) {
	link := readFixture(t, "retribution_no_settings_link.txt")
	imported, err := Import(link)
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if got, want := imported.Character.Class, proto.Class_ClassPaladin.String(); got != want {
		t.Fatalf("class = %q, want %q", got, want)
	}
	if got, want := imported.Character.Spec, "RetributionPaladin"; got != want {
		t.Fatalf("spec = %q, want %q", got, want)
	}
	if imported.Character.Phase != 0 || imported.Character.Iterations != 0 || imported.Character.FixedRngSeed {
		t.Fatalf("summary = %+v, want zero-valued missing simulation settings", imported.Character)
	}
}

func TestImportRejectsRaidAndMalformedLinks(t *testing.T) {
	initFixtures(t)
	validLink := readFixture(t, "fixed_individual_link.txt")
	parts := bytes.Split([]byte(validLink), []byte("#"))
	fragment := string(parts[1])

	testCases := []struct {
		name string
		link string
		code string
	}{
		{"no fragment", "https://wowsims.com/tbc/mage/", "invalid_link"},
		{"empty fragment", "https://wowsims.com/tbc/mage/#", "invalid_link"},
		{"exclamation fragment", "https://wowsims.com/tbc/mage/#!", "invalid_link"},
		{"invalid base64", "https://wowsims.com/tbc/mage/#notbase64!", "invalid_link"},
		{"raid link", "https://wowsims.com/tbc/raid/#" + fragment, "unsupported_link"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Import(tc.link)
			if err == nil {
				t.Fatalf("Import(%q) succeeded unexpectedly", tc.link)
			}
			var valErr *ValidationError
			if !errors.As(err, &valErr) {
				t.Fatalf("err is not ValidationError: %v", err)
			}
			if valErr.Code != tc.code {
				t.Fatalf("error code = %q, want %q (msg: %s)", valErr.Code, tc.code, valErr.Message)
			}
		})
	}
}

func TestNewRequestDoesNotMutateImportedSettings(t *testing.T) {
	imported := mustImportFixture(t)
	before := mustMarshal(t, imported.Settings)
	req := imported.NewRequest(1234)
	if req == nil {
		t.Fatal("NewRequest returned nil")
	}
	if req.SimOptions.Iterations != 1234 {
		t.Fatalf("iterations = %d, want 1234", req.SimOptions.Iterations)
	}
	if after := mustMarshal(t, imported.Settings); !bytes.Equal(before, after) {
		t.Fatal("NewRequest mutated imported settings")
	}
}
func unsupportedReferenceCases() []struct {
	name   string
	mutate func(*proto.ItemSpec)
	code   string
} {
	return []struct {
		name   string
		mutate func(*proto.ItemSpec)
		code   string
	}{
		{"item", func(i *proto.ItemSpec) { i.Id = 999999 }, "incompatible_item"},
		{"random suffix", func(i *proto.ItemSpec) { i.RandomSuffix = 999999 }, "incompatible_random_suffix"},
		{"gem", func(i *proto.ItemSpec) { i.Gems = []int32{999999} }, "incompatible_gem"},
		{"enchant", func(i *proto.ItemSpec) { i.Enchant = 999999 }, "incompatible_enchant"},
	}
}

func testImportRejectsUnsupportedReferences(t *testing.T, selectItem func(*proto.IndividualSimSettings) *proto.ItemSpec) {
	t.Helper()
	for _, tc := range unsupportedReferenceCases() {
		t.Run(tc.name, func(t *testing.T) {
			settings := googleProto.Clone(fixtureIndividualSettings()).(*proto.IndividualSimSettings)
			tc.mutate(selectItem(settings))
			link, err := encodeIndividualLink(settings)
			if err != nil {
				t.Fatalf("failed to encode test link: %v", err)
			}

			_, err = Import(link)
			if err == nil {
				t.Fatal("Import succeeded unexpectedly")
			}
			var valErr *ValidationError
			if !errors.As(err, &valErr) {
				t.Fatalf("err is not ValidationError: %v", err)
			}
			if valErr.Code != tc.code {
				t.Fatalf("error code = %q, want %q (msg: %s)", valErr.Code, tc.code, valErr.Message)
			}
		})
	}
}

func TestImportRejectsUnsupportedEquipmentReferences(t *testing.T) {
	testImportRejectsUnsupportedReferences(t, func(settings *proto.IndividualSimSettings) *proto.ItemSpec {
		return settings.Player.Equipment.Items[0]
	})
}

func TestImportRejectsUnsupportedItemSwapReferences(t *testing.T) {
	testImportRejectsUnsupportedReferences(t, func(settings *proto.IndividualSimSettings) *proto.ItemSpec {
		settings.Player.EnableItemSwap = true
		settings.Player.ItemSwap = &proto.ItemSwap{
			Items: []*proto.ItemSpec{{Id: 24266}},
		}
		return settings.Player.ItemSwap.Items[0]
	})
}
