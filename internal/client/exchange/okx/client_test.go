package okx

import (
	"testing"
)

type resp struct {
	Code string `json:"code"`
	Data []struct {
		AcctLv              string        `json:"acctLv"`
		AcctStpMode         string        `json:"acctStpMode"`
		AutoLoan            bool          `json:"autoLoan"`
		CtIsoMode           string        `json:"ctIsoMode"`
		EnableSpotBorrow    bool          `json:"enableSpotBorrow"`
		GreeksType          string        `json:"greeksType"`
		FeeType             string        `json:"feeType"`
		Ip                  string        `json:"ip"`
		Type                string        `json:"type"`
		KycLv               string        `json:"kycLv"`
		Label               string        `json:"label"`
		Level               string        `json:"level"`
		LevelTmp            string        `json:"levelTmp"`
		LiquidationGear     string        `json:"liquidationGear"`
		MainUid             string        `json:"mainUid"`
		MgnIsoMode          string        `json:"mgnIsoMode"`
		OpAuth              string        `json:"opAuth"`
		Perm                string        `json:"perm"`
		PosMode             string        `json:"posMode"`
		RoleType            string        `json:"roleType"`
		SpotBorrowAutoRepay bool          `json:"spotBorrowAutoRepay"`
		SpotOffsetType      string        `json:"spotOffsetType"`
		SpotRoleType        string        `json:"spotRoleType"`
		SpotTraderInsts     []interface{} `json:"spotTraderInsts"`
		StgyType            string        `json:"stgyType"`
		TraderInsts         []interface{} `json:"traderInsts"`
		Uid                 string        `json:"uid"`
		SettleCcy           string        `json:"settleCcy"`
		SettleCcyList       []string      `json:"settleCcyList"`
	} `json:"data"`
	Msg string `json:"msg"`
}

func TestNewOkxClient(t *testing.T) {

}
