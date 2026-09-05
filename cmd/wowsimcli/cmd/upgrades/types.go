package upgrades

import (
	"context"

	"github.com/wowsims/tbc/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

const (
	SimulatorRevision = "88fb853466a391e731e12de012f6707a11e94446"
	DatabaseRevision  = "84555ed6e3ddf19edca22204892a2366e1a177da"
)

type CharacterSummary struct {
	Name             string             `json:"name"`
	Class            string             `json:"class"`
	Spec             string             `json:"spec"`
	Race             string             `json:"race"`
	EquippedItems    int                `json:"equippedItems"`
	Professions      []proto.Profession `json:"professions"`
	Phase            int32              `json:"phase"`
	Iterations       int32              `json:"iterations"`
	FixedRngSeed     bool               `json:"fixedRngSeed"`
	EncounterTargets int                `json:"encounterTargets"`
}

type ImportedSettings struct {
	Link             string                       `json:"link"`
	Settings         *proto.IndividualSimSettings `json:"-"`
	SettingsDigest   string                       `json:"settingsDigest"`
	Character        CharacterSummary             `json:"character"`
	SimulatorVersion string                       `json:"simulatorVersion"`
	DatabaseVersion  string                       `json:"databaseVersion"`
}

type ArmoryData struct {
	Gear         []GearSlotData     `json:"gear"`
	Stats        map[string]float64 `json:"stats"`
	DerivedStats map[string]float64 `json:"derivedStats"`
}

type GearSlotData struct {
	Slot         proto.ItemSlot     `json:"slot"`
	SlotName     string             `json:"slotName"`
	ItemID       int32              `json:"itemId"`
	ItemName     string             `json:"itemName"`
	Quality      proto.ItemQuality  `json:"quality"`
	Icon         string             `json:"icon"`
	Phase        int32              `json:"phase"`
	Ilvl         int32              `json:"ilvl"`
	SetName      string             `json:"setName"`
	Stats        map[string]float64 `json:"stats"`
	RandomSuffix *RandomSuffixData  `json:"randomSuffix"`
	Sockets      []SocketData       `json:"sockets"`
	SocketBonus  SocketBonusData    `json:"socketBonus"`
	Enchant      *EnchantData       `json:"enchant"`
}

type RandomSuffixData struct {
	ID    int32              `json:"id"`
	Name  string             `json:"name"`
	Stats map[string]float64 `json:"stats"`
}

type GemData struct {
	ID    int32              `json:"id"`
	Name  string             `json:"name"`
	Icon  string             `json:"icon"`
	Color proto.GemColor     `json:"color"`
	Stats map[string]float64 `json:"stats"`
}

type SocketData struct {
	Color proto.GemColor `json:"color"`
	Gem   *GemData       `json:"gem"`
}

type SocketBonusData struct {
	Stats  map[string]float64 `json:"stats"`
	Active bool               `json:"active"`
}

type EnchantData struct {
	ID          int32              `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Icon        string             `json:"icon"`
	Stats       map[string]float64 `json:"stats"`
}

type ValidationError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

func cloneMessage[M googleProto.Message](m M) M {
	if googleProto.Message(m) == nil {
		var zero M
		return zero
	}
	return googleProto.Clone(m).(M)
}

// cloneOrEmpty clones m, or returns empty when m is nil.
func cloneOrEmpty[M googleProto.Message](m M, empty M) M {
	if googleProto.Message(m) == nil {
		return empty
	}
	return cloneMessage(m)
}

type ContentFilters struct {
	MaxPhase       int32                      `json:"maxPhase"`
	SourceKinds    []proto.SourceFilterOption `json:"sourceKinds"`
	SourceNames    []string                   `json:"sourceNames"`
	IncludeUnknown bool                       `json:"includeUnknown"`
}

type ItemPolicy struct {
	GemBySocket   map[proto.GemColor]int32 `json:"gemBySocket"`
	MaxGemQuality proto.ItemQuality        `json:"maxGemQuality"`
	EnchantByType map[proto.ItemType]int32 `json:"enchantByType"`
}

type ExclusionSummary struct {
	UnknownSource int            `json:"unknownSource"`
	Source        int            `json:"source"`
	Compatibility int            `json:"compatibility"`
	Policy        int            `json:"policy"`
	Reasons       map[string]int `json:"reasons"`
}

type UIItemSummary struct {
	ID      int32             `json:"id"`
	Name    string            `json:"name"`
	Icon    string            `json:"icon"`
	Quality proto.ItemQuality `json:"quality"`
	Phase   int32             `json:"phase"`
	Type    proto.ItemType    `json:"type"`
	Slot    proto.ItemSlot    `json:"slot"`
}

type SourceSummary struct {
	Kind     proto.SourceFilterOption `json:"kind"`
	Name     string                   `json:"name"`
	Zone     string                   `json:"zone"`
	Category string                   `json:"category"`
}

type PolicyApplication struct {
	GemIDs        []int32 `json:"gemIds"`
	EnchantID     int32   `json:"enchantId"`
	SocketChoices []int32 `json:"socketChoices"`
}

type PolicyError struct {
	Reason string
}

func (e *PolicyError) Error() string {
	return e.Reason
}

type Candidate struct {
	Item       UIItemSummary         `json:"item"`
	TargetSlot proto.ItemSlot        `json:"targetSlot"`
	Displaced  []UIItemSummary       `json:"displaced"`
	Request    *proto.RaidSimRequest `json:"-"`
	Applied    PolicyApplication     `json:"applied"`
	Source     SourceSummary         `json:"source"`
}

type BuildResult struct {
	Candidates []Candidate      `json:"candidates"`
	Excluded   ExclusionSummary `json:"excluded"`
}

type SimulationOptions struct {
	ScreeningIterations    int32 `json:"screeningIterations"`
	ConfirmationIterations int32 `json:"confirmationIterations"`
}

type RankRequest struct {
	Imported *ImportedSettings `json:"imported"`
	Filters  ContentFilters    `json:"filters"`
	Policy   ItemPolicy        `json:"policy"`
	Options  SimulationOptions `json:"options"`
}

type ConfirmedUpgrade struct {
	Rank                 int               `json:"rank"`
	Item                 UIItemSummary     `json:"item"`
	TargetSlot           proto.ItemSlot    `json:"targetSlot"`
	Displaced            []UIItemSummary   `json:"displaced"`
	Source               SourceSummary     `json:"source"`
	Applied              PolicyApplication `json:"applied"`
	DpsDelta             float64           `json:"dpsDelta"`
	RelativeGainPercent  float64           `json:"relativeGainPercent"`
	StandardError        float64           `json:"standardError"`
	ConfidenceInterval95 [2]float64        `json:"confidenceInterval95"`
	Iterations           int32             `json:"iterations"`
	TooCloseToCall       bool              `json:"tooCloseToCall"`
	Assumptions          ReportAssumptions `json:"assumptions"`
}

type FailedCandidate struct {
	Item       UIItemSummary  `json:"item"`
	TargetSlot proto.ItemSlot `json:"targetSlot"`
	Reason     string         `json:"reason"`
}

type ReportAssumptions struct {
	LinkDigest             string           `json:"linkDigest"`
	MaxPhase               int32            `json:"maxPhase"`
	SourceKinds            []string         `json:"sourceKinds"`
	SourceNames            []string         `json:"sourceNames"`
	IncludeUnknown         bool             `json:"includeUnknown"`
	MaxGemQuality          string           `json:"maxGemQuality"`
	GemBySocket            map[string]int32 `json:"gemBySocket"`
	EnchantByType          map[string]int32 `json:"enchantByType"`
	ScreeningIterations    int32            `json:"screeningIterations"`
	ConfirmationIterations int32            `json:"confirmationIterations"`
}

type BaselineSummary struct {
	Dps        float64 `json:"dps"`
	DpsStdev   float64 `json:"dpsStdev"`
	Iterations int32   `json:"iterations"`
}

type UpgradeReport struct {
	Baseline               BaselineSummary    `json:"baseline"`
	Character              CharacterSummary   `json:"character"`
	Confirmed              []ConfirmedUpgrade `json:"confirmed"`
	Excluded               ExclusionSummary   `json:"excluded"`
	Failed                 []FailedCandidate  `json:"failed"`
	Assumptions            ReportAssumptions  `json:"assumptions"`
	AssumptionsFingerprint string             `json:"assumptionsFingerprint"`
	SimulatorRevision      string             `json:"simulatorRevision"`
	DatabaseRevision       string             `json:"databaseRevision"`
	CapTruncatedTieRegion  bool               `json:"capTruncatedTieRegion"`
}

type Progress struct {
	Stage     string `json:"stage"`
	Completed int32  `json:"completed"`
	Total     int32  `json:"total"`
}

type DPSResult struct {
	Average    float64
	Stdev      float64
	Iterations int32
}

type Simulator interface {
	Run(ctx context.Context, request *proto.RaidSimRequest, onProgress func(completed, total int32)) (DPSResult, error)
}
