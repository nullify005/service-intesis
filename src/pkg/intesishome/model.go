package intesishome

const (
	_defaultHostname string = "https://user.intesishome.com"
	ControlEndpoint  string = "/api.php/get/control"
	StatusCommand    string = `{"status":{"hash":"x"},"config":{"hash":"x"}}`
	StatusVersion    string = "1.8.5"
	WriteCommand     string = `{"command":"set","data":{"deviceId":%v,"uid":%v,"value":%v,"seqNo":0}}`
	WriteAuth        string = `{"command":"connect_req","data":{"token":%v}}`
	CommandConnOk    string = "connect_rsp"
	CommandSetAck    string = "set_ack"
	CommandAuthOk    string = "ok"
)

type IntesisHome struct {
	username   string
	password   string
	hostname   string
	mock       bool
	serverIP   string
	serverPort int
	token      int
}

type ControlResponse struct {
	Config struct {
		Token          int     `json:"token"`
		PushToken      string  `json:"pushToken"`
		LastAppVersion string  `json:"lastAppVersion"`
		ForceUpdate    int     `json:"forceUpdate"`
		SetDelay       float64 `json:"setDelay"`
		ServerIP       string  `json:"serverIP"`
		ServerPort     int     `json:"serverPort"`
		Hash           string  `json:"hash"`
		Inst           []struct {
			ID      int      `json:"id"`
			Order   int      `json:"order"`
			Name    string   `json:"name"`
			Devices []Device `json:"devices"`
		} `json:"inst"`
	} `json:"config"`
	Status struct {
		Hash   string `json:"hash"`
		Status []struct {
			DeviceID int64 `json:"deviceId"`
			UID      int   `json:"uid"`
			Value    int   `json:"value"`
			Name     string
		} `json:"status"`
	} `json:"status"`
	ErrorCode    int    `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

type Device struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	FamilyID       int    `json:"familyId"`
	ModelID        int    `json:"modelId"`
	InstallationID int    `json:"installationId"`
	ZoneID         int    `json:"zoneId"`
	Order          int    `json:"order"`
	Widgets        []int  `json:"widgets"`
}

type CommandResponse struct {
	Command string `json:"command"`
	Data    struct {
		DeviceID int64  `json:"devceId"`
		SeqNo    int    `json:"seqNo"`
		Rssi     int    `json:"rssi"`
		Status   string `json:"status"`
	} `json:"data"`
}

type CommandRequest struct {
	Command string `json:"command"`
	Data    struct {
		DeviceID int64 `json:"deviceId,omitempty"`
		Uid      int   `json:"uid,omitempty"`
		Value    int   `json:"value,omitempty"`
		SeqNo    int   `json:"seqNo,omitempty"`
		Token    int   `json:"token,omitempty"`
	} `json:"data"`
}
