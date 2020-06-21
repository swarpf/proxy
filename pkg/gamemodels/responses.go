package gamemodels

import "reflect"

//
// type: ApiResponse
type ApiResponse struct {
	Command     string `json:"command"`
	RetCode     int    `json:"ret_code"`
	TsVal       int    `json:"ts_val"`
	TValue      int    `json:"tvalue"`
	TValueLocal int    `json:"tvaluelocal"`
	TZone       string `json:"tzone"`
}

//
// type(ApiResponse): GetWizardInfo
type GetWizardInfo struct {
	ApiResponse
	WizardInfo WizardInfo `json:"wizard_info"`
}

//
// type(ApiResponse): GetWizardInfo
type BattleDungeonStart struct {
	ApiResponse
	WizardInfo WizardInfo `json:"wizard_info"`
	BattleKey  int        `json:"battle_key"`
}

//
// type(ApiResponse): GetWizardInfo
type BattleDungeonResultV2 struct {
	ApiResponse
	WinLose         bool                          `json:"win_lose"`
	WizardInfo      WizardInfo                    `json:"wizard_info"`
	Reward          DungeonReward                 `json:"reward"`
	ChangedItemList []DungeonChangedItemListEntry `json:"changed_item_list"`
}

//
// type(ApiResponse): GetWizardInfo
type BattleTrialTowerStartV2 struct {
	ApiResponse
	WizardInfo WizardInfo `json:"wizard_info"`
}

//
// type(ApiResponse): GetWizardInfo
type BattleTrialTowerResultV2 struct {
	ApiResponse
	WinLose    bool          `json:"win_lose"`
	WizardInfo WizardInfo    `json:"wizard_info"`
	Reward     DungeonReward `json:"reward"`
	FloorId    int           `json:"floor_id"`
}

//
// fixme(lyrex): i'm pretty sure this should not be a thing
func CommandToType(command string) reflect.Type {
	commandTypeMap := map[string]reflect.Type{
		"GetWizardInfo":             reflect.TypeOf(GetWizardInfo{}),
		"BattleDungeonStart":        reflect.TypeOf(BattleDungeonStart{}),
		"BattleDungeonResult_V2":    reflect.TypeOf(BattleDungeonResultV2{}),
		"BattleTrialTowerStart_v2":  reflect.TypeOf(BattleTrialTowerStartV2{}),
		"BattleTrialTowerResult_v2": reflect.TypeOf(BattleTrialTowerResultV2{}),
	}

	t, ok := commandTypeMap[command]
	if !ok {
		return reflect.Type(nil)
	}

	return t
}
