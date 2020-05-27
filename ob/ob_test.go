package ob

import (
	"encoding/hex"
	"flag"
	"testing"
)

func TestComet(t *testing.T) {
	var c Comet
	hex.Decode(c[:], []byte("1fe49fba73b5725bdb99f904e7acf465"))
	if c.String() != "~mirryc-patpex-saldef-padhep--nactyr-sovsev-mosber-bonwet" {
		t.Fatal("bad comet name:", c.String())
	} else if c.Parent().String() != "~bonwet" {
		t.Fatal("wrong comet parent:", c.Parent().String())
	}

	sk, _ := hex.DecodeString("4230e39bc7a387ec3b9f4f08d68c0ea0093e0bb4ef1f5c618494ae6c7fb4f4e2c2604f698a5e28a63996eb6886d04816188d538864883083d98fa549be5e5bfc55")
	jam := formatUW(jamComet(c, sk))
	if jam != "0w2.G~ySL.nOjiN.-P1C4.gOh2D.6z0IA.q4cQt.sIsQN.gLhji.DI65N.uBE~J.Btagz.2K3~v.q1pY4.Q0t6q.MgDPV.TSgZ7.zPv6o.8g7w0.svYA~.tetGV.buTc~.89PRD.EO-M1" {
		t.Fatal("bad jam for comet")
	}
}

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
