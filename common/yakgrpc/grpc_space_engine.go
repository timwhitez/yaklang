package yakgrpc

import (
	"context"
	_ "embed"
	uuid "github.com/satori/go.uuid"
	"github.com/yaklang/yaklang/common/consts"
	"github.com/yaklang/yaklang/common/utils/spacengine"
	"github.com/yaklang/yaklang/common/utils/spacengine/fofa"
	"github.com/yaklang/yaklang/common/utils/spacengine/zoomeye"
	"github.com/yaklang/yaklang/common/yak"
	"github.com/yaklang/yaklang/common/yak/antlr4yak"
	"github.com/yaklang/yaklang/common/yak/yaklib"
	"github.com/yaklang/yaklang/common/yakgrpc/ypb"
)

const (
	SPACE_ENGINE_ZOOMEYE = "zoomeye"
	SPACE_ENGINE_FOFA    = "fofa"
	SPACE_ENGINE_SHODAN  = "shodan"
	SPACE_ENGINE_HUNTER  = "hunter"

	SPACE_ENGINE_STATUS_NORMAL       = "normal"
	SPACE_ENGINE_STATUS_ERROR        = "error"
	SPACE_ENGINE_STATUS_EMPTY_KEY    = "empty_key"
	SPACE_ENGINE_STATUS_INVALID_TYPE = "invalid_type"
)

func (s *Server) GetSpaceEngineStatus(ctx context.Context, req *ypb.GetSpaceEngineStatusRequest) (*ypb.SpaceEngineStatus, error) {
	//var status = SPACE_ENGINE_STATUS_NORMAL
	//info := "ZoomEye额度按月刷新"
	var status = SPACE_ENGINE_STATUS_INVALID_TYPE
	var info = ""
	var raw []byte
	var remain int64
	switch req.GetType() {
	case SPACE_ENGINE_ZOOMEYE:
		status = SPACE_ENGINE_STATUS_NORMAL
		info = "ZoomEye额度按月刷新"
		key := consts.GetThirdPartyApplicationConfig("zoomeye").APIKey
		if key == "" {
			status = SPACE_ENGINE_STATUS_EMPTY_KEY
			info = "ZoomEye API Key为空"
			break
		}
		result, err := zoomeye.ZoomeyeUserProfile(consts.GetThirdPartyApplicationConfig("zoomeye").APIKey)
		if err != nil {
			status = SPACE_ENGINE_STATUS_ERROR
			info = err.Error()
			break
		}
		// res := result.Get(`resources`)
		quota := result.Get("quota_info")
		remain = quota.Get("remain_free_quota").Int() + quota.Get("remain_pay_quota").Int()
		raw = []byte(result.Raw)
	case SPACE_ENGINE_SHODAN:
		status = SPACE_ENGINE_STATUS_NORMAL
		info = "账户正常"
		key := consts.GetThirdPartyApplicationConfig("shodan").APIKey
		if key == "" {
			status = SPACE_ENGINE_STATUS_EMPTY_KEY
			info = "Shodan API Key为空"
			break
		}
		result, err := spacengine.ShodanUserProfile(key)
		if err != nil {
			status = SPACE_ENGINE_STATUS_ERROR
			info = err.Error()
			break
		}
		_ = result
		remain = -1
	case SPACE_ENGINE_HUNTER:
		status = SPACE_ENGINE_STATUS_NORMAL
		info = "Hunter免费额度按月刷新"
		key := consts.GetThirdPartyApplicationConfig("hunter").APIKey
		if key == "" {
			status = SPACE_ENGINE_STATUS_EMPTY_KEY
			info = "Hunter API Key为空"
			break
		}
	case SPACE_ENGINE_FOFA:
		status = SPACE_ENGINE_STATUS_NORMAL
		info = "普通账户"
		key := consts.GetThirdPartyApplicationConfig("fofa").APIKey
		if key == "" {
			status = SPACE_ENGINE_STATUS_EMPTY_KEY
			info = "FOFA API Key 为空"
			break
		}
		email := consts.GetThirdPartyApplicationConfig("fofa").UserIdentifier
		if email == "" {
			status = SPACE_ENGINE_STATUS_EMPTY_KEY
			info = "FOFA Email 为空"
			break
		}
		client := fofa.NewFofaClient(email, key)
		user, err := client.UserInfo()
		if user.Vip {
			info = "VIP账户"
		}
		if err != nil {
			status = SPACE_ENGINE_STATUS_ERROR
			info = err.Error()
			break
		}
		remain = user.RemainApiQuery
	default:
		status = SPACE_ENGINE_STATUS_INVALID_TYPE
	}
	return &ypb.SpaceEngineStatus{
		Type:   req.GetType(),
		Status: status,
		Info:   info,
		Raw:    raw,
		Remain: remain,
	}, nil
}

//go:embed grpc_space_engine.yak
var spaceEngineExecCode string

func (s *Server) FetchPortAssetFromSpaceEngine(req *ypb.FetchPortAssetFromSpaceEngineRequest, stream ypb.Yak_FetchPortAssetFromSpaceEngineServer) error {
	engine := yak.NewYakitVirtualClientScriptEngine(yaklib.NewVirtualYakitClient(stream.Send))
	runtimeId := uuid.NewV4().String()
	engine.RegisterEngineHooks(func(engine *antlr4yak.Engine) error {
		engine.SetVar("FILTER", req.GetFilter())
		engine.SetVar("SCAN_VERIFY", req.GetScanBeforeSave())
		engine.SetVar("TOTAL_PAGE", req.GetMaxPage())
		engine.SetVar("ENGINE_TYPE", req.GetType())
		engine.SetVar("CONCURRENT", req.GetConcurrent())
		yak.BindYakitPluginContextToEngine(engine, &yak.YakitPluginContext{
			PluginName: "space-engine",
			RuntimeId:  runtimeId,
			Proxy:      req.GetProxy(),
		})
		return nil
	})
	return engine.ExecuteWithContext(stream.Context(), spaceEngineExecCode)
}
