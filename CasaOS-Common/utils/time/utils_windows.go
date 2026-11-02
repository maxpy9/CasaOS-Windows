//go:build windows

package time

import (
	"golang.org/x/sys/windows/registry"
)

// windowsToIANA maps the time zone names of Windows to the names used by Linux
// containers. It holds the default zone of each Windows time zone.
var windowsToIANA = map[string]string{
	"Afghanistan Standard Time":       "Asia/Kabul",
	"Alaskan Standard Time":           "America/Anchorage",
	"Arab Standard Time":              "Asia/Riyadh",
	"Arabian Standard Time":           "Asia/Dubai",
	"Arabic Standard Time":            "Asia/Baghdad",
	"Argentina Standard Time":         "America/Buenos_Aires",
	"Atlantic Standard Time":          "America/Halifax",
	"AUS Central Standard Time":       "Australia/Darwin",
	"AUS Eastern Standard Time":       "Australia/Sydney",
	"Azerbaijan Standard Time":        "Asia/Baku",
	"Azores Standard Time":            "Atlantic/Azores",
	"Canada Central Standard Time":    "America/Regina",
	"Cape Verde Standard Time":        "Atlantic/Cape_Verde",
	"Caucasus Standard Time":          "Asia/Yerevan",
	"Cen. Australia Standard Time":    "Australia/Adelaide",
	"Central America Standard Time":   "America/Guatemala",
	"Central Asia Standard Time":      "Asia/Almaty",
	"Central Brazilian Standard Time": "America/Cuiaba",
	"Central Europe Standard Time":    "Europe/Budapest",
	"Central European Standard Time":  "Europe/Warsaw",
	"Central Pacific Standard Time":   "Pacific/Guadalcanal",
	"Central Standard Time":           "America/Chicago",
	"Central Standard Time (Mexico)":  "America/Mexico_City",
	"China Standard Time":             "Asia/Shanghai",
	"E. Africa Standard Time":         "Africa/Nairobi",
	"E. Australia Standard Time":      "Australia/Brisbane",
	"E. Europe Standard Time":         "Europe/Chisinau",
	"E. South America Standard Time":  "America/Sao_Paulo",
	"Eastern Standard Time":           "America/New_York",
	"Egypt Standard Time":             "Africa/Cairo",
	"Fiji Standard Time":              "Pacific/Fiji",
	"FLE Standard Time":               "Europe/Kiev",
	"Georgian Standard Time":          "Asia/Tbilisi",
	"GMT Standard Time":               "Europe/London",
	"Greenland Standard Time":         "America/Godthab",
	"Greenwich Standard Time":         "Atlantic/Reykjavik",
	"GTB Standard Time":               "Europe/Bucharest",
	"Hawaiian Standard Time":          "Pacific/Honolulu",
	"India Standard Time":             "Asia/Calcutta",
	"Iran Standard Time":              "Asia/Tehran",
	"Israel Standard Time":            "Asia/Jerusalem",
	"Jordan Standard Time":            "Asia/Amman",
	"Korea Standard Time":             "Asia/Seoul",
	"Mauritius Standard Time":         "Indian/Mauritius",
	"Middle East Standard Time":       "Asia/Beirut",
	"Morocco Standard Time":           "Africa/Casablanca",
	"Mountain Standard Time":          "America/Denver",
	"Myanmar Standard Time":           "Asia/Rangoon",
	"Namibia Standard Time":           "Africa/Windhoek",
	"Nepal Standard Time":             "Asia/Katmandu",
	"New Zealand Standard Time":       "Pacific/Auckland",
	"Newfoundland Standard Time":      "America/St_Johns",
	"North Asia Standard Time":        "Asia/Krasnoyarsk",
	"Pacific SA Standard Time":        "America/Santiago",
	"Pacific Standard Time":           "America/Los_Angeles",
	"Pakistan Standard Time":          "Asia/Karachi",
	"Romance Standard Time":           "Europe/Paris",
	"Russian Standard Time":           "Europe/Moscow",
	"SA Eastern Standard Time":        "America/Cayenne",
	"SA Pacific Standard Time":        "America/Bogota",
	"SA Western Standard Time":        "America/La_Paz",
	"SE Asia Standard Time":           "Asia/Bangkok",
	"Singapore Standard Time":         "Asia/Singapore",
	"South Africa Standard Time":      "Africa/Johannesburg",
	"Sri Lanka Standard Time":         "Asia/Colombo",
	"Taipei Standard Time":            "Asia/Taipei",
	"Tokyo Standard Time":             "Asia/Tokyo",
	"Turkey Standard Time":            "Europe/Istanbul",
	"Ulaanbaatar Standard Time":       "Asia/Ulaanbaatar",
	"US Eastern Standard Time":        "America/Indianapolis",
	"US Mountain Standard Time":       "America/Phoenix",
	"UTC":                             "Etc/UTC",
	"Venezuela Standard Time":         "America/Caracas",
	"Vladivostok Standard Time":       "Asia/Vladivostok",
	"W. Australia Standard Time":      "Australia/Perth",
	"W. Central Africa Standard Time": "Africa/Lagos",
	"W. Europe Standard Time":         "Europe/Berlin",
	"West Asia Standard Time":         "Asia/Tashkent",
	"West Pacific Standard Time":      "Pacific/Port_Moresby",
	"Yakutsk Standard Time":           "Asia/Yakutsk",
}

var systemTimeZoneName string

// GetSystemTimeZoneName returns the IANA name of the time zone of the computer.
func GetSystemTimeZoneName() string {
	if systemTimeZoneName == "" {
		key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\TimeZoneInformation`, registry.QUERY_VALUE)
		if err != nil {
			println("cannot read the time zone from the registry: " + err.Error())
			return ""
		}
		defer key.Close()

		name, _, err := key.GetStringValue("TimeZoneKeyName")
		if err != nil {
			println("cannot read the time zone from the registry: " + err.Error())
			return ""
		}

		iana, ok := windowsToIANA[name]
		if !ok {
			println("unknown time zone " + name)
			return ""
		}

		systemTimeZoneName = iana
	}

	return systemTimeZoneName
}
