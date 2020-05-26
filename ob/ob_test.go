package ob

import (
	"flag"
	"testing"
)

func TestPoints(t *testing.T) {
	nec := AzimuthPoint(1)
	if !nec.IsGalaxy() {
		t.Error("nec should be a galaxy")
	} else if nec.String() != "~nec" {
		t.Error("wrong name for point", nec.String())
	} else if ap, err := PointFromName(nec.String()); err != nil || ap != nec {
		t.Error("bad point from name:", uint32(ap), err)
	}

	marnec := nec.ChildStar(1)
	if !marnec.IsStar() {
		t.Error("marcnec should be a star")
	} else if marnec.Parent() != nec {
		t.Error("marnec's parent should be nec")
	} else if marnec.String() != "~marnec" {
		t.Error("wrong name for point", marnec.String())
	} else if ap, err := PointFromName(marnec.String()); err != nil || ap != marnec {
		t.Error("bad point from name:", uint32(ap), err)
	}

	ralnyt := marnec.ChildPlanet(1)
	if !ralnyt.IsPlanet() {
		t.Error("ralnyt-botdyt should be a planet")
	} else if ralnyt.Parent() != marnec {
		t.Error("ralnyt-botdyt's parent should be marnec")
	} else if ralnyt.String() != "~ralnyt-botdyt" {
		t.Error("wrong name for point", ralnyt.String())
	} else if ap, err := PointFromName(ralnyt.String()); err != nil || ap != ralnyt {
		t.Error("bad point from name:", uint32(ap), err)
	}
}

var testInjective = flag.Bool("injective", false, "run the injectivity test")

func TestInjectivity(t *testing.T) {
	if !*testInjective {
		t.Skip("skipping injectivity test")
	}
	for i := AzimuthPoint(1); i != 0; i++ {
		if i%50e6 == 0 {
			println(i)
		}
		if fynd(fein(i)) != i {
			t.Fatal("patp not injective")
		}
	}
}
