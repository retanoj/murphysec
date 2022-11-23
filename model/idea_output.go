package model

import (
	"context"
	"encoding/json"
	"github.com/murphysecurity/murphysec/utils"
	"github.com/murphysecurity/murphysec/utils/must"
)

func GenerateIdeaErrorOutput(e error) string {
	iec := GetIdeaErrCode(e)
	return string(must.A(json.Marshal(struct {
		ErrCode IdeaErrCode `json:"err_code"`
		ErrMsg  string      `json:"err_msg"`
	}{ErrCode: iec, ErrMsg: e.Error()})))
}

type PluginOutput struct {
	//ProjectName string       `json:"project_name"`
	SubtaskName string       `json:"subtask_name"`
	Username    string       `json:"username"`
	ErrCode     IdeaErrCode  `json:"err_code"`
	IssuesCount int          `json:"issues_count,omitempty"`
	Comps       []PluginComp `json:"comps,omitempty"`
	//ProjectScore     int          `json:"project_score"`
	//SurpassScore     string       `json:"surpass_score"`
	IssuesLevelCount struct {
		Critical int `json:"critical,omitempty"`
		High     int `json:"high,omitempty"`
		Medium   int `json:"medium,omitempty"`
		Low      int `json:"low,omitempty"`
	} `json:"issues_level_count,omitempty"`
	TaskId    string `json:"task_id"`
	SubtaskId string `json:"subtask_id"`
	//TotalContributors int            `json:"total_contributors"`
	//ProjectId         string         `json:"project_id"`
	InspectErrors     []InspectError `json:"inspect_errors,omitempty"`
	DependenciesCount int            `json:"dependencies_count"`
	//InspectReportUrl  string         `json:"inspect_report_url"`
}

type PluginComp struct {
	CompName        string `json:"comp_name"`
	ShowLevel       int    `json:"show_level"`
	MinFixedVersion string `json:"min_fixed_version"`
	//DisposePlan        PluginCompFixList    `json:"dispose_plan"`
	Vulns       []PluginVulnDetailInfo `json:"vulns"`
	CompVersion string                 `json:"version"`
	//License            PluginCompLicense         `json:"license,omitempty"`
	Licenses           []LicenseItem `json:"licenses"`
	Solutions          []Solution    `json:"solutions,omitempty"`
	IsDirectDependency bool          `json:"is_direct_dependency"`
	//Language           string        `json:"language"`
	//FixType            string        `json:"fix_type"`
	CompSecScore  int         `json:"comp_sec_score"`
	FixPlanList   FixPlanList `json:"fix_plan_list"`
	DependentPath []string    `json:"dependent_path"`
}

type PluginCompSolution struct {
	Compatibility *int   `json:"compatibility,omitempty"`
	Description   string `json:"description"`
	Type          string `json:"type,omitempty"`
}

func GetIDEAOutput(ctx context.Context) PluginOutput {
	var task = UseScanTask(ctx)
	var r = task.result
	var pluginOutput = PluginOutput{
		SubtaskName: r.SubtaskName,
		Comps:       make([]PluginComp, 0),
		IssuesCount: r.LeakNum,
		IssuesLevelCount: struct {
			Critical int `json:"critical,omitempty"`
			High     int `json:"high,omitempty"`
			Medium   int `json:"medium,omitempty"`
			Low      int `json:"low,omitempty"`
		}{
			Critical: r.CriticalNum,
			High:     r.HighNum,
			Medium:   r.MediumNum,
			Low:      r.LowNum,
		},
		TaskId:            r.TaskId,
		SubtaskId:         r.SubtaskId,
		DependenciesCount: r.RelyNum,
	}

	var vulnListMapper = func(effects []ScanResultCompEffect) (rs []PluginVulnDetailInfo) {
		for _, effect := range effects {
			info, ok := r.VulnInfoMap[effect.MpsId]
			if !ok {
				continue // skip item if detailed information not found
			}
			var d = PluginVulnDetailInfo{
				MpsId:           info.MpsID,
				CveId:           info.CveID,
				Description:     info.Description,
				Level:           info.Level,
				Influence:       info.Influence,
				Poc:             info.Poc,
				PublishTime:     int(info.PublishedDate.Unix()),
				AffectedVersion: effect.EffectVersion,
				MinFixedVersion: effect.MinFixedVersion,
				References:      utils.NoNilSlice(info.ReferenceURLList),
				Solutions:       utils.NoNilSlice(effect.Solutions),
				SuggestLevel:    info.FixSuggestionLevel,
				Title:           info.Title,
			}
			rs = append(rs, d)
		}
		return
	}

	for _, comp := range r.CompInfoList {
		var pc = PluginComp{
			CompName:           comp.CompName,
			CompVersion:        comp.CompVersion,
			ShowLevel:          comp.SuggestLevel,
			MinFixedVersion:    comp.MinFixedVersion,
			Vulns:              utils.NoNilSlice(vulnListMapper(comp.VulnList)),
			Licenses:           utils.NoNilSlice(comp.LicenseList),
			Solutions:          utils.NoNilSlice(comp.Solutions),
			IsDirectDependency: comp.IdDirectDependency,
			CompSecScore:       comp.CompSecScore,
			FixPlanList:        comp.FixPlanList,
			DependentPath:      utils.NoNilSlice(comp.DependentPath),
		}
		pluginOutput.Comps = append(pluginOutput.Comps, pc)
	}
	return pluginOutput
}

type PluginVulnDetailInfo struct {
	MpsId           string         `json:"mps_id"`
	CveId           string         `json:"cve_id"`
	Description     string         `json:"description"`
	Level           string         `json:"level"`
	Influence       int            `json:"influence"`
	Poc             bool           `json:"poc"`
	PublishTime     int            `json:"publish_time"`
	AffectedVersion string         `json:"affected_version"`
	MinFixedVersion string         `json:"min_fixed_version"`
	References      []ReferenceURL `json:"references"`
	Solutions       []Solution     `json:"solutions"`
	SuggestLevel    string         `json:"suggest_level"`
	Title           string         `json:"title"`
}
