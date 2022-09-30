package intesishome

const ApiEndpoint string = "https://user.intesishome.com/api.php/get/control"
const StatusCommand string = `{"status":{"hash":"x"},"config":{"hash":"x"}}`
const StatusVersion string = "1.8.5"

type Connection struct {
	Username string
	Password string
	Endpoint string
	Mock     bool
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
