package providers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/bitly/oauth2_proxy/api"
)

type DingTalkProvider struct {
	*ProviderData
	Departments     map[int64]string
	CorpAccessToken struct {
		CorpID        string
		CorpSSOSecret string
		AccessToken   string
		ExpiresOn     time.Time
	}
}

func NewDingTalkProvider(p *ProviderData) *DingTalkProvider {
	p.ProviderName = "DingTalk"
	if p.LoginURL == nil || p.LoginURL.String() == "" {
		p.LoginURL = &url.URL{
			Scheme: "https",
			Host:   "oapi.dingtalk.com",
			Path:   "/connect/qrconnect",
		}
	}
	if p.RedeemURL == nil || p.RedeemURL.String() == "" {
		p.RedeemURL = &url.URL{
			Scheme: "https",
			Host:   "oapi.dingtalk.com",
			Path:   "/login/oauth/access_token",
		}
	}
	// ValidationURL is the API Base URL
	if p.ValidateURL == nil || p.ValidateURL.String() == "" {
		p.ValidateURL = &url.URL{
			Scheme: "https",
			Host:   "oapi.dingtalk.com",
			Path:   "/",
		}
	}
	if p.Scope == "" {
		p.Scope = "snsapi_login"
	}
	return &DingTalkProvider{ProviderData: p}
}

func (p *DingTalkProvider) GetLoginURL(redirectURI, state string) string {
	var a url.URL
	a = *p.LoginURL
	params, _ := url.ParseQuery(a.RawQuery)
	params.Set("redirect_uri", redirectURI)
	params.Add("scope", p.Scope)
	params.Set("appid", p.ClientID)
	params.Set("response_type", "code")
	params.Add("state", state)
	a.RawQuery = params.Encode()
	return a.String()
}

/**
Get access token of app via client id and client secret
https://open-doc.dingtalk.com/microapp/serverapi2/athq8o#%E8%8E%B7%E5%8F%96%E9%92%89%E9%92%89%E5%BC%80%E6%94%BE%E5%BA%94%E7%94%A8%E7%9A%84access_token
**/
func (p *DingTalkProvider) GetAppAccessToken() (accessToken string, err error) {
	params := url.Values{
		"appid":     {p.ClientID},
		"appsecret": {p.ClientSecret},
	}
	gettokenURL := &url.URL{
		Scheme:   p.RedeemURL.Scheme,
		Host:     p.RedeemURL.Host,
		Path:     "/sns/gettoken",
		RawQuery: params.Encode(),
	}

	var req *http.Request
	req, err = http.NewRequest("GET", gettokenURL.String(), nil)
	if err != nil {
		return
	}

	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("got %d from %q %s", resp.StatusCode, gettokenURL.String(), body)
		return
	}

	// blindly try json and x-www-form-urlencoded
	var jsonResponse struct {
		AccessToken string `json:"access_token"`
		ErrorCode   int    `json:"errcode"`
		ErrMessage  string `json:"errmsg"`
	}
	err = json.Unmarshal(body, &jsonResponse)
	if err == nil {
		if 0 == jsonResponse.ErrorCode {
			accessToken = jsonResponse.AccessToken
			return
		}
		err = fmt.Errorf("got error code %d from %q %s", jsonResponse.ErrorCode, gettokenURL.String(), jsonResponse.ErrMessage)
		return
	}

	err = fmt.Errorf("no access token of app found %s", body)
	return
}

/**
Get persistent code then sns token as session state
https://open-doc.dingtalk.com/microapp/serverapi2/athq8o#%E8%8E%B7%E5%8F%96%E7%94%A8%E6%88%B7%E6%8E%88%E6%9D%83%E7%9A%84%E6%8C%81%E4%B9%85%E6%8E%88%E6%9D%83%E7%A0%81
*/
func (p *DingTalkProvider) GetSNSToken(accessToken string, code string) (s *SessionState, err error) {
	params := url.Values{}
	params.Add("access_token", accessToken)
	getPersistentCodeURL := &url.URL{
		Scheme:   p.RedeemURL.Scheme,
		Host:     p.RedeemURL.Host,
		Path:     "/sns/get_persistent_code",
		RawQuery: params.Encode(),
	}
	body := map[string]string{"tmp_auth_code": code}
	var req *http.Request
	jsonBody, err := json.Marshal(body)
	req, err = http.NewRequest("POST", getPersistentCodeURL.String(), bytes.NewBuffer(jsonBody))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	var persistentCodeResp struct {
		PersistentCode string `json:"persistent_code"`
		Openid         string `json:"openid"`
		Unionid        string `json:"unionid"`
		ErrorCode      int    `json:"errcode"`
		ErrMessage     string `json:"errmsg"`
	}
	err = api.RequestJson(req, &persistentCodeResp)
	if err != nil {
		return nil, err
	}
	if persistentCodeResp.ErrorCode != 0 {
		err = fmt.Errorf("got error code %d from %q %s", persistentCodeResp.ErrorCode, getPersistentCodeURL.String(), persistentCodeResp.ErrMessage)
		return
	}

	getSNSTokenURL := &url.URL{
		Scheme:   p.RedeemURL.Scheme,
		Host:     p.RedeemURL.Host,
		Path:     "/sns/get_sns_token",
		RawQuery: params.Encode(),
	}
	body = map[string]string{"openid": persistentCodeResp.Openid,
		"persistent_code": persistentCodeResp.PersistentCode}
	jsonBody, err = json.Marshal(body)
	req, err = http.NewRequest("POST", getSNSTokenURL.String(), bytes.NewBuffer(jsonBody))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	var snsTokenResp struct {
		SNSToken   string `json:"sns_token"`
		ExpiresIn  int    `json:"expires_in"`
		ErrorCode  int    `json:"errcode"`
		ErrMessage string `json:"errmsg"`
	}
	err = api.RequestJson(req, &snsTokenResp)
	if err != nil {
		return nil, err
	}
	if snsTokenResp.ErrorCode != 0 {
		err = fmt.Errorf("got error code %d from %q %s", snsTokenResp.ErrorCode, getSNSTokenURL.String(), snsTokenResp.ErrMessage)
		return
	}
	s = &SessionState{
		AccessToken: snsTokenResp.SNSToken,
		ExpiresOn:   time.Now().Add(time.Duration(snsTokenResp.ExpiresIn-60) * time.Second).Truncate(time.Second),
	}

	return
}

func (p *DingTalkProvider) Redeem(redirectURL, code string) (s *SessionState, err error) {
	if code == "" {
		err = errors.New("missing code")
		return
	}

	accessToken, err := p.GetAppAccessToken()
	if err != nil {
		return
	}

	s, err = p.GetSNSToken(accessToken, code)
	return
}

func (p *DingTalkProvider) SetCorpInfoAndDepartments(corpID, corpSSOSecret string, departments []string) (err error) {
	p.CorpAccessToken.CorpID = corpID
	p.CorpAccessToken.CorpSSOSecret = corpSSOSecret
	err = p.RefreshCorpAccessToken()
	if err != nil {
		return err
	}
	departmentMap := make(map[string]string)
	for _, department := range departments {
		departmentMap[department] = ""
	}
	return p.InitialDepartments(departmentMap)
}

func (p *DingTalkProvider) InitialDepartments(departments map[string]string) (err error) {
	if len(departments) > 0 {
		params := url.Values{}
		params.Add("access_token", p.CorpAccessToken.AccessToken)
		params.Add("fetch_child", "true")
		getDepartmentsURL := &url.URL{
			Scheme:   p.RedeemURL.Scheme,
			Host:     p.RedeemURL.Host,
			Path:     "/department/list",
			RawQuery: params.Encode(),
		}
		var req *http.Request
		req, err = http.NewRequest("GET", getDepartmentsURL.String(), nil)
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		var departmentListResp struct {
			ErrorCode   int    `json:"errcode"`
			ErrMessage  string `json:"errmsg"`
			Departments []struct {
				ID       int64  `json:"id"`
				Name     string `json:"name"`
				ParentID int64  `json:"parentid"`
			} `json:"department"`
		}
		err = api.RequestJson(req, &departmentListResp)
		if err != nil {
			return
		}
		if departmentListResp.ErrorCode != 0 {
			err = fmt.Errorf("got error code %d from %q %s", departmentListResp.ErrorCode, getDepartmentsURL.String(), departmentListResp.ErrMessage)
			return
		}
		departmentsInfo := make(map[int64]string)
		p.Departments = make(map[int64]string)
		sort.Slice(departmentListResp.Departments, func(i, j int) bool {
			if departmentListResp.Departments[i].ID == 1 {
				return true
			}
			if departmentListResp.Departments[j].ID == 1 {
				return false
			}
			if departmentListResp.Departments[i].ParentID == 1 {
				return true
			}
			if departmentListResp.Departments[i].ParentID == departmentListResp.Departments[j].ID {
				return false
			}
			if departmentListResp.Departments[j].ParentID == departmentListResp.Departments[i].ID {
				return true
			}
			return departmentListResp.Departments[i].ID < departmentListResp.Departments[j].ID
		})
		for _, department := range departmentListResp.Departments {
			if department.ID == 1 {
				departmentsInfo[department.ID] = department.Name
			} else {
				departmentsInfo[department.ID] = departmentsInfo[department.ParentID] + "/" + department.Name
			}
			_, exist := departments[departmentsInfo[department.ID]]
			if exist {
				p.Departments[department.ID] = departmentsInfo[department.ID]
			} else if _, parentExist := p.Departments[department.ParentID]; parentExist {
				p.Departments[department.ID] = departmentsInfo[department.ID]
			}
		}
		if len(departments) > len(p.Departments) {
			log.Printf("Incorrect departments whitelist %v in remote departments %#v", departments, departmentsInfo)
		} else {
			log.Printf("Prepremitted departments %v", p.Departments)
		}
	}
	return
}

func (p *DingTalkProvider) RefreshCorpAccessToken() (err error) {
	params := url.Values{}
	params.Add("appkey", p.CorpAccessToken.CorpID)
	params.Add("appsecret", p.CorpAccessToken.CorpSSOSecret)
	getCorpAccessTokenURL := &url.URL{
		Scheme:   p.RedeemURL.Scheme,
		Host:     p.RedeemURL.Host,
		Path:     "/gettoken",
		RawQuery: params.Encode(),
	}
	var req *http.Request
	req, err = http.NewRequest("GET", getCorpAccessTokenURL.String(), nil)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	var corpAccessTokenResp struct {
		AccessToken string `json:"access_token"`
		ErrorCode   int    `json:"errcode"`
		ErrMessage  string `json:"errmsg"`
	}
	err = api.RequestJson(req, &corpAccessTokenResp)
	if err != nil {
		return err
	}
	if corpAccessTokenResp.ErrorCode != 0 {
		err = fmt.Errorf("got error code %d from %q %s", corpAccessTokenResp.ErrorCode, getCorpAccessTokenURL.String(), corpAccessTokenResp.ErrMessage)
		return
	}
	p.CorpAccessToken.AccessToken = corpAccessTokenResp.AccessToken
	p.CorpAccessToken.ExpiresOn = time.Now().Add(time.Duration(7200-60) * time.Second).Truncate(time.Second)
	return nil
}

func (p *DingTalkProvider) GetEmailAddress(s *SessionState) (string, error) {
	/**
	Get union id of user
	https://open-doc.dingtalk.com/microapp/serverapi2/athq8o#a-nameafurtga%E8%8E%B7%E5%8F%96%E7%94%A8%E6%88%B7%E6%8E%88%E6%9D%83%E7%9A%84%E4%B8%AA%E4%BA%BA%E4%BF%A1%E6%81%AF
	*/
	params := url.Values{}
	params.Add("sns_token", s.AccessToken)
	getSNSUserURL := &url.URL{
		Scheme:   p.RedeemURL.Scheme,
		Host:     p.RedeemURL.Host,
		Path:     "/sns/getuserinfo",
		RawQuery: params.Encode(),
	}

	req, err := http.NewRequest("GET", getSNSUserURL.String(), nil)
	if err != nil {
		return "", fmt.Errorf("could not create new GET request: %v", err)
	}

	var snsUserResp struct {
		ErrorCode  int    `json:"errcode"`
		ErrMessage string `json:"errmsg"`
		User       struct {
			Nick    string `json:"nick"`
			OpenID  string `json:"openid"`
			UnionID string `json:"unionid"`
		} `json:"user_info"`
	}
	err = api.RequestJson(req, &snsUserResp)
	if err != nil {
		return "", err
	}
	if snsUserResp.ErrorCode != 0 {
		err = fmt.Errorf("got error code %d from %q %s", snsUserResp.ErrorCode, getSNSUserURL.String(), snsUserResp.ErrMessage)
		return "", err
	}

	/**
	Get user id of user in corp
	https://open-doc.dingtalk.com/microapp/serverapi2/ege851#a-names9a%E6%A0%B9%E6%8D%AEunionid%E8%8E%B7%E5%8F%96userid
	*/
	params = url.Values{}
	if p.CorpAccessToken.ExpiresOn.Before(time.Now()) {
		err = p.RefreshCorpAccessToken()
		if err != nil {
			return "", err
		}
	}
	params.Add("access_token", p.CorpAccessToken.AccessToken)
	params.Add("unionid", snsUserResp.User.UnionID)
	getUserIDByUnionIDURL := &url.URL{
		Scheme:   p.RedeemURL.Scheme,
		Host:     p.RedeemURL.Host,
		Path:     "/user/getUseridByUnionid",
		RawQuery: params.Encode(),
	}

	req, err = http.NewRequest("GET", getUserIDByUnionIDURL.String(), nil)
	if err != nil {
		return "", fmt.Errorf("could not create new GET request: %v", err)
	}

	var userIDResp struct {
		ErrorCode   int    `json:"errcode"`
		ErrMessage  string `json:"errmsg"`
		ContactType int    `json:"contactType"`
		UserID      string `json:"userid"`
	}
	err = api.RequestJson(req, &userIDResp)
	if err != nil {
		return "", err
	}
	if userIDResp.ErrorCode != 0 {
		err = fmt.Errorf("got error code %d from %q %s", userIDResp.ErrorCode, getUserIDByUnionIDURL.String(), userIDResp.ErrMessage)
		return "", err
	}

	/**
	Get user info in corp
	https://open-doc.dingtalk.com/microapp/serverapi2/ege851#a-names1a%E8%8E%B7%E5%8F%96%E7%94%A8%E6%88%B7%E8%AF%A6%E6%83%85
	*/
	params = url.Values{}
	params.Add("access_token", p.CorpAccessToken.AccessToken)
	params.Add("userid", userIDResp.UserID)
	getUserInfoURL := &url.URL{
		Scheme:   p.RedeemURL.Scheme,
		Host:     p.RedeemURL.Host,
		Path:     "/user/get",
		RawQuery: params.Encode(),
	}

	req, err = http.NewRequest("GET", getUserInfoURL.String(), nil)
	if err != nil {
		return "", fmt.Errorf("could not create new GET request: %v", err)
	}

	var userResp struct {
		ErrorCode  int     `json:"errcode"`
		ErrMessage string  `json:"errmsg"`
		UserID     string  `json:"userid"`
		Name       string  `json:"name"`
		Email      string  `json:"email"`
		OrgEmail   string  `json:"orgEmail"`
		Mobile     string  `json:"mobile"`
		Department []int64 `json:"department"`
	}
	err = api.RequestJson(req, &userResp)
	if err != nil {
		return "", err
	}
	if userIDResp.ErrorCode != 0 {
		err = fmt.Errorf("got error code %d from %q %s", userResp.ErrorCode, getUserInfoURL.String(), userResp.ErrMessage)
		return "", err
	}

	if p.Departments == nil || p.hasDepartment(userResp.Department) {
		s.User = userResp.Name
		s.Email = userResp.OrgEmail
		if s.Email == "" {
			s.Email = userResp.Email
		}
	}

	return s.Email, nil
}

func (p *DingTalkProvider) hasDepartment(departments []int64) bool {
	for _, departID := range departments {
		if _, exist := p.Departments[departID]; exist {
			return true
		}
	}
	log.Printf("Missing Department in white list:%v in %#v", departments, p.Departments)
	return false
}
