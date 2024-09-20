package route

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRoute(t *testing.T) {
	sid := "SID"
	minLevel := uint64(100)
	maxLevel := uint64(200)
	routeSegment := "Segment"
	star := "STAR"
	noteIDs := []uint64{1, 2, 3}

	route := NewRoute("ADEP", &sid, &minLevel, &maxLevel, routeSegment, &star, "ADES", noteIDs)

	require.Equal(t, "ADEP", route.ADEPOrEntry(), "expected ADEPOrEntry to be 'ADEP'")
	require.Equal(t, sid, *route.SID(), "expected SID to be 'SID'")
	require.Equal(t, minLevel, *route.MinLevel(), "expected MinLevel to be '100'")
	require.Equal(t, maxLevel, *route.MaxLevel(), "expected MaxLevel to be '200'")
	require.Equal(t, routeSegment, route.RouteSegment(), "expected RouteSegment to be 'Segment'")
	require.Equal(t, star, *route.STAR(), "expected STAR to be 'STAR'")
	require.Equal(t, "ADES", route.ADESOrExit(), "expected ADESOrExit to be 'ADES'")
	require.Equal(t, noteIDs, route.NoteIDs(), "expected NoteIDs to be '[1, 2, 3]'")
}
