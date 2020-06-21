package gamemodels

import (
	"errors"
	"fmt"
)

//
// enum: GameItem
type GameItem int

const (
	CategoryMonster           GameItem = 1
	CategoryCurrency          GameItem = 6
	CategoryRune              GameItem = 8
	CategorySummonScroll      GameItem = 9
	CategoryBooster           GameItem = 10
	CategoryEssence           GameItem = 11
	CategoryMonsterPiece      GameItem = 12
	CateogryGuildMonsterPiece GameItem = 19
	CategoryRainbowmon        GameItem = 25
	CategoryRuneCraft         GameItem = 27
	CategoryCraftStuff        GameItem = 29
	CategorySecretDungeon     GameItem = 30
	CategoryMaterialMonster   GameItem = 61
)

func (g GameItem) String() string {
	names := map[GameItem]string{
		CategoryMonster:           "Monster",
		CategoryCurrency:          "Currency",
		CategoryRune:              "Rune",
		CategorySummonScroll:      "SummonScroll",
		CategoryBooster:           "Booster",
		CategoryEssence:           "Essence",
		CategoryMonsterPiece:      "MonsterPiece",
		CateogryGuildMonsterPiece: "GuildMonsterPiece",
		CategoryRainbowmon:        "Rainbowmon",
		CategoryRuneCraft:         "RuneCraft",
		CategoryCraftStuff:        "CraftStuff",
		CategorySecretDungeon:     "SecretDungeon",
		CategoryMaterialMonster:   "MaterialMonster",
	}

	name, ok := names[g]
	if !ok {
		return "Unknown"
	}

	return name
}

//
// enum: UnitAttribute
type UnitAttribute int

const (
	AttributeWater UnitAttribute = 1
	AttributeFire  UnitAttribute = 2
	AttributeWind  UnitAttribute = 3
	AttributeLight UnitAttribute = 4
	AttributeDark  UnitAttribute = 5
)

func (ua UnitAttribute) String() string {
	names := map[UnitAttribute]string{
		AttributeWater: "Water",
		AttributeFire:  "Fire",
		AttributeWind:  "Wind",
		AttributeLight: "Light",
		AttributeDark:  "Dark",
	}

	name, ok := names[ua]
	if !ok {
		return "Unknown"
	}

	return name
}

//
// enum: EffectType
type EffectType int

const (
	Hp     EffectType = 1
	HpPct  EffectType = 2
	Atk    EffectType = 3
	AtkPct EffectType = 4
	Def    EffectType = 5
	DefPct EffectType = 6
	Spd    EffectType = 8
	Cr     EffectType = 9
	Cd     EffectType = 10
	Res    EffectType = 11
	Acc    EffectType = 12
)

func (et EffectType) String() string {
	names := map[EffectType]string{
		Hp:     "HP",
		HpPct:  "HP%",
		Atk:    "ATK",
		AtkPct: "ATK%",
		Def:    "DEF",
		DefPct: "DEF%",
		Spd:    "SPD",
		Cr:     "CR",
		Cd:     "CD",
		Res:    "RES",
		Acc:    "ACC",
	}

	name, ok := names[et]
	if !ok {
		return "Unknown"
	}

	return name
}

//
// enum: RuneSet
type RuneSet int

const (
	Energy        RuneSet = 1
	Guard         RuneSet = 2
	Swift         RuneSet = 3
	Blade         RuneSet = 4
	Rage          RuneSet = 5
	Focus         RuneSet = 6
	Endure        RuneSet = 7
	Fatal         RuneSet = 8
	Despair       RuneSet = 10
	Vampire       RuneSet = 11
	Violent       RuneSet = 13
	Nemesis       RuneSet = 14
	Will          RuneSet = 15
	Shield        RuneSet = 16
	Revenge       RuneSet = 17
	Destroy       RuneSet = 18
	Fight         RuneSet = 19
	Determination RuneSet = 20
	Enhance       RuneSet = 21
	Accuracy      RuneSet = 22
	Tolerance     RuneSet = 23
)

func (rs RuneSet) String() string {
	names := map[RuneSet]string{
		Energy:        "Energy",
		Guard:         "Guard",
		Swift:         "Swift",
		Blade:         "Blade",
		Rage:          "Rage",
		Focus:         "Focus",
		Endure:        "Endure",
		Fatal:         "Fatal",
		Despair:       "Despair",
		Vampire:       "Vampire",
		Violent:       "Violent",
		Nemesis:       "Nemesis",
		Will:          "Will",
		Shield:        "Shield",
		Revenge:       "Revenge",
		Destroy:       "Destroy",
		Fight:         "Fight",
		Determination: "Determination",
		Enhance:       "Enhance",
		Accuracy:      "Accuracy",
		Tolerance:     "Tolerance",
	}

	name, ok := names[rs]
	if !ok {
		return "Unknown"
	}

	return name
}

//
// enum: RuneQuality
type RuneQuality int

const (
	Common RuneQuality = 1
	Magic  RuneQuality = 2
	Rare   RuneQuality = 3
	Hero   RuneQuality = 4
	Legend RuneQuality = 5
)

func (rq RuneQuality) String() string {
	names := map[RuneQuality]string{
		Common: "Common",
		Magic:  "Magic",
		Rare:   "Rare",
		Hero:   "Hero",
		Legend: "Legend",
	}

	name, ok := names[rq]
	if !ok {
		return "Unknown"
	}

	return name
}

//
// type: Building
type Building struct {
	BuildingId       int     `json:"building_id"`
	WizardId         int     `json:"wizard_id"`
	IslandId         int     `json:"island_id"`
	BuildingMasterId int     `json:"building_master_id"`
	PosX             int     `json:"pos_x"`
	PosY             int     `json:"pos_y"`
	GainPerHour      float32 `json:"gain_per_hour"`
	HarvestMax       *int    `json:"harvest_max"`
	HarvestAvailable *int    `json:"harvest_available"`
	NextHarvest      *int    `json:"next_harvest"`
}

func (b Building) Equal(other Building) bool {
	return b.BuildingId == other.BuildingId
}

//
// type: WizardInfo
type WizardInfo struct {
	WizardId           int     `json:"wizard_id"`
	WizardName         string  `json:"wizard_name"`
	WizardMana         int     `json:"wizard_mana"`
	WizardCrystal      int     `json:"wizard_crystal"`
	WizardLevel        int     `json:"wizard_level"`
	WizardEnergy       int     `json:"wizard_energy"`
	EnergyMax          int     `json:"energy_max"`
	EnergyPerMin       float32 `json:"energy_per_min"`
	NextEnergyGain     int     `json:"next_energy_gain"`
	PvpEvent           bool    `json:"pvp_event"`
	MailBoxEvent       bool    `json:"mail_box_event"`
	SocialPointCurrent int     `json:"social_point_current"`
	SocialPointMax     int     `json:"social_point_max"` // default: 3000
}

func (w WizardInfo) Equal(other WizardInfo) bool {
	return w.WizardId == other.WizardId
}

//
// type: DungeonReward
type DungeonReward struct {
	Mana       int         `json:"mana"`
	Crystal    int         `json:"crystal"`
	Energy     int         `json:"energy"`
	Crate      interface{} `json:"crate"`
	EventCrate interface{} `json:"event_crate"`
}

//
// type: DungeonChangedItemListEntry
type DungeonChangedItemListInfoEntry map[string]interface{}
type DungeonChangedViewListInfoEntry map[string]interface{}
type DungeonChangedItemListEntry struct {
	Type GameItem                        `json:"type"`
	Info DungeonChangedItemListInfoEntry `json:"info"`
	View DungeonChangedViewListInfoEntry `json:"view"`
}

//
// type: RuneStat
type RuneStat struct {
	EffectType  EffectType `json:"effect_type"`
	EffectValue int        `json:"effect_value"`
	IsEnchanted bool       `json:"is_enchanted"`
	GrindValue  int        `json:"grind_value"`
}

func (r RuneStat) IsGrinded() bool {
	return r.GrindValue != 0
}

// todo(lyrex): figure out how to write custom tojson/fromjson functions

func NewRuneStatFromObject(obj interface{}) (*RuneStat, error) {
	if obj == nil {
		return nil, errors.New("obj is nil")
	}
	data, ok := obj.([]int)

	if !ok {
		return nil, errors.New("obj is not a rune stat")
	}

	dataLen := len(data)
	if dataLen < 2 || (dataLen != 2 && dataLen != 4) {
		return nil, fmt.Errorf("obj has an invalid length: %d", dataLen)
	}

	// return nil if all values are 0. this means it's an empty stat
	// todo(lyrex): check if this is needed
	if data[0] == 0 || data[1] == 0 {
		return nil, nil
	}

	effectType := EffectType(data[0])
	effectValue := data[1]

	var runeStat *RuneStat
	switch dataLen {
	case 2:
		runeStat = &RuneStat{
			EffectType:  effectType,
			EffectValue: effectValue,
			IsEnchanted: false,
			GrindValue:  0,
		}
		break
	case 4:
		runeStat = &RuneStat{
			EffectType:  effectType,
			EffectValue: effectValue,
			IsEnchanted: data[2] == 1,
			GrindValue:  data[3],
		}
		break
	default:
		return nil, errors.New("obj data is corrupt")
	}

	return runeStat, nil
}

//
// type: Rune

type Rune struct {
	RuneId          int         `json:"rune_id"`
	WizardId        int         `json:"wizard_id"`
	RuneSet         RuneSet     `json:"set_id"`
	Stars           int         `json:"class"`
	Level           int         `json:"upgrade_curr"`
	Slot            int         `json:"slot_no"`
	Quality         RuneQuality `json:"rank"`
	OriginalQuality RuneQuality `json:"extra"`
	Ancient         bool        `json:"ancient"`
	SellValue       int         `json:"sell_value"`
	MainStat        RuneStat    `json:"pri_eff"`
	InnateStat      *RuneStat   `json:"prefix_eff"`
	Substats        []RuneStat  `json:"sec_eff"`
}

/*

@dataclass
class Rune:
    rune_id: int
    wizard_id: int
    rune_set: RuneSet

    stars: int
    level: int
    slot: int
    quality: RuneQuality
    original_quality: RuneQuality
    ancient: bool
    sell_value: int
    main_stat: RuneStat
    innate_stat: Optional[RuneStat]
    substats: List[RuneStat]

    @classmethod
    def from_json_object(cls, rune_data: Dict):
        ok, key = cls._validate_fields(rune_data)
        if not ok:
            raise ValueError(f"Rune is missing required field {key}")

        rune_id = rune_data["rune_id"]
        wizard_id = rune_data["wizard_id"]
        rune_set = RuneSet(rune_data["set_id"])
        stars = rune_data["class"]
        if stars > 10:
            stars -= 10
            ancient = True
        else:
            ancient = False
        level = rune_data["upgrade_curr"]
        slot = rune_data["slot_no"]
        quality = RuneQuality(rune_data["rank"])
        original_quality = RuneQuality(rune_data["extra"])
        sell_value = rune_data["sell_value"]

        main_stat = RuneStat.from_object(rune_data["pri_eff"])
        innate_stat = RuneStat.from_object(rune_data["prefix_eff"])
        substats = [RuneStat.from_object(x) for x in rune_data["sec_eff"]]

        r = cls(rune_id, wizard_id, rune_set, stars, level, slot, quality, original_quality, ancient, sell_value,
                main_stat, innate_stat, substats)

        return r

    @staticmethod
    def _validate_fields(rune_data: Dict) -> Tuple[bool, Optional[str]]:
        _REQUIRED_KEYS = ["rune_id", "wizard_id", "set_id", "class", "upgrade_curr", "slot_no", "rank",
                          "extra", "sell_value", "pri_eff", "prefix_eff", "sec_eff"]
        for key in _REQUIRED_KEYS:
            if key not in rune_data:
                return False, key
        return True, None


@dataclass_json
@dataclass
class Unit:
    unit_id: int
    wizard_id: int
    unit_master_id: int
    unit_level: int
    unit_class: int = field(metadata=config(field_name="class"))
    unit_con: int = field(metadata=config(field_name="con"))
    unit_atk: int = field(metadata=config(field_name="atk"))
    unit_def: int = field(metadata=config(field_name="def"))
    unit_spd: int = field(metadata=config(field_name="spd"))
    unit_resist: int = field(metadata=config(field_name="resist"))
    unit_accuracy: int = field(metadata=config(field_name="accuracy"))
    unit_critical_rate: int = field(metadata=config(field_name="critical_rate"))
    unit_critical_damage: int = field(metadata=config(field_name="critical_damage"))
    # unit_skills: List[UnitSkill] = field(metadata=config(field_name="skills"))
    unit_runes: List[Rune] = field(metadata=config(field_name="runes", decoder=Rune.from_json_object))
    unit_attribute: UnitAttribute = field(metadata=config(field_name="attribute"))

    def __eq__(self, other: object) -> bool:
        if not isinstance(other, Unit):
            return False
        return self.unit_id == other.unit_id
*/
