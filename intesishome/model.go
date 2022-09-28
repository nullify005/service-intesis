package intesishome

const ApiEndpoint string = "https://user.intesishome.com/api.php/get/control"
const StatusCommand string = `{"status":{"hash":"x"},"config":{"hash":"x"}}`
const StatusVersion string = "1.8.5"
const mockPayload string = "static/ControlResponse.json"
const mappingPayload string = "static/StateMapping.json"

type Connection struct {
	Username string
	Password string
	Mock     string
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

type DeviceState struct {
	Power       uint
	Mode        uint
	FanSpeed    int
	Temperature float32
	SetPoint    float32
}

var Capability = map[int]string{
	1:     "power",
	2:     "mode",
	3:     "unknown",
	4:     "fan_speed",
	5:     "vvane",
	6:     "hvane",
	7:     "unknown",
	9:     "setpoint",
	10:    "temperature",
	12:    "remote_controller_lock",
	13:    "working_hours",
	14:    "alarm_status",
	15:    "error_code",
	34:    "quiet_mode",
	35:    "setpoint_min",
	36:    "setpoint_max",
	37:    "outdoor_temp",
	38:    "water_outlet_temperature",
	39:    "water_inlet_temperature",
	42:    "climate_working_mode",
	44:    "tank_working_mode",
	45:    "tank_water_temperature",
	46:    "solar_status",
	48:    "thermoshift_heat_eco",
	49:    "thermoshift_cool_eco",
	50:    "thermoshift_heat_powerful",
	51:    "thermoshift_cool_powerful",
	52:    "thermoshift_tank_eco",
	53:    "thermoshift_tank_powerful",
	54:    "error_reset",
	55:    "heat_thermo_shift",
	56:    "cool_water_setpoint_temperature",
	57:    "tank_setpoint_temperature",
	58:    "operating_mode",
	60:    "heat_8_10",
	61:    "config_mode_map",
	62:    "runtime_mode_restrictions",
	63:    "config_horizontal_vanes",
	64:    "config_vertical_vanes",
	65:    "config_quiet",
	66:    "config_confirm_off",
	67:    "config_fan_map",
	68:    "instant_power_consumption",
	69:    "accumulated_power_consumption",
	75:    "config_operating_mode",
	77:    "config_vanes_pulse",
	80:    "aquarea_tank_consumption",
	81:    "aquarea_cool_consumption",
	82:    "aquarea_heat_consumption",
	83:    "heat_high_water_set_temperature",
	84:    "heating_off_temperature",
	87:    "heater_setpoint_temperature",
	90:    "water_target_temperature",
	95:    "heat_interval",
	107:   "aquarea_working_hours",
	123:   "ext_thermo_control",
	124:   "tank_present",
	125:   "solar_priority",
	134:   "heat_low_outdoor_set_temperature",
	135:   "heat_high_outdoor_set_temperature",
	136:   "heat_low_water_set_temperature",
	137:   "farenheit_type",
	140:   "extremes_protection_status",
	144:   "error_code",
	148:   "extremes_protection",
	149:   "binary_input",
	153:   "config_binary_input",
	168:   "uid_binary_input_on_off",
	169:   "uid_binary_input_occupancy",
	170:   "uid_binary_input_window",
	181:   "mainenance_w_reset",
	182:   "mainenance_wo_reset",
	183:   "filter_clean",
	184:   "filter_due_hours",
	185:   "uid_185",
	186:   "uid_186",
	191:   "uid_binary_input_sleep_mode",
	192:   "error_address",
	50000: "external_led",
	50001: "internal_led",
	50002: "internal_temperature_offset",
	50003: "temp_limitation",
	50004: "cool_temperature_min",
	50005: "cool_temperature_max",
	50006: "heat_temperature_min",
	50007: "heat_temperature_min",
	60002: "rssi",
}
