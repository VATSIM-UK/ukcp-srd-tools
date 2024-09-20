package route

import "encoding/json"

// A route is a single route row in the SRD file.
type Route struct {
	departureAirfieldOrEntryPoint string
	standardInstrumentDeparture   *string
	minimumFlightLevel            *uint64
	maximumFlightLevel            *uint64
	routeSegment                  string
	standardTerminalArrivalRoute  *string
	arrivalAirfieldOrExitPoint    string
	noteIDs                       []uint64
}

// NewRoute creates a new route, it is unvalidated.
func NewRoute(
	ADEPOrEntry string,
	SID *string,
	minLevel *uint64,
	maxLevel *uint64,
	routeSegment string,
	STAR *string,
	ADESOrExit string,
	noteIDs []uint64,
) *Route {
	return &Route{
		departureAirfieldOrEntryPoint: ADEPOrEntry,
		standardInstrumentDeparture:   SID,
		minimumFlightLevel:            minLevel,
		maximumFlightLevel:            maxLevel,
		routeSegment:                  routeSegment,
		standardTerminalArrivalRoute:  STAR,
		arrivalAirfieldOrExitPoint:    ADESOrExit,
		noteIDs:                       noteIDs,
	}
}

func (r *Route) ADEPOrEntry() string {
	return r.departureAirfieldOrEntryPoint
}

func (r *Route) SID() *string {
	return r.standardInstrumentDeparture
}

func (r *Route) MinLevel() *uint64 {
	return r.minimumFlightLevel
}

func (r *Route) MaxLevel() *uint64 {
	return r.maximumFlightLevel
}

func (r *Route) RouteSegment() string {
	return r.routeSegment
}

func (r *Route) STAR() *string {
	return r.standardTerminalArrivalRoute
}

func (r *Route) ADESOrExit() string {
	return r.arrivalAirfieldOrExitPoint
}

func (r *Route) NoteIDs() []uint64 {
	return r.noteIDs
}

// ToJSON converts the Route struct to a JSON string.
// The properties are marshalled manually using an inline struct to avoid exposing the struct fields.
func (r *Route) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(struct {
		DepartureAirfieldOrEntryPoint string   `json:"departure_airfield_or_entry_point"`
		StandardInstrumentDeparture   *string  `json:"standard_instrument_departure"`
		MinimumFlightLevel            *uint64  `json:"minimum_flight_level"`
		MaximumFlightLevel            *uint64  `json:"maximum_flight_level"`
		RouteSegment                  string   `json:"route_segment"`
		StandardTerminalArrivalRoute  *string  `json:"standard_terminal_arrival_route"`
		ArrivalAirfieldOrExitPoint    string   `json:"arrival_airfield_or_exit_point"`
		NoteIDs                       []uint64 `json:"note_ids"`
	}{
		DepartureAirfieldOrEntryPoint: r.departureAirfieldOrEntryPoint,
		StandardInstrumentDeparture:   r.standardInstrumentDeparture,
		MinimumFlightLevel:            r.minimumFlightLevel,
		MaximumFlightLevel:            r.maximumFlightLevel,
		RouteSegment:                  r.routeSegment,
		StandardTerminalArrivalRoute:  r.standardTerminalArrivalRoute,
		ArrivalAirfieldOrExitPoint:    r.arrivalAirfieldOrExitPoint,
		NoteIDs:                       r.noteIDs,
	})
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
