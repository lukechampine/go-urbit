package atom

import (
	"math/big"
	"testing"
)

func TestFormat(t *testing.T) {
	testCases := []struct {
		dec string
		exp map[Aura]string
	}{
		{
			dec: "0",
			exp: map[Aura]string{
				"dr":  "~s0",
				"p":   "~zod",
				"s":   "--0",
				"sb":  "--0b0",
				"sv":  "--0v0",
				"sw":  "--0w0",
				"sx":  "--0x0",
				"t":   "''",
				"ta":  "~.",
				"tas": "%$",
				"u":   "0",
				"ub":  "0b0",
				"uv":  "0v0",
				"uw":  "0w0",
				"ux":  "0x0",
			},
		},
		{
			dec: "56",
			exp: map[Aura]string{
				"dr":  "~s0..0000.0000.0000.0038",
				"p":   "~bes",
				"s":   "--28",
				"sb":  "--0b1.1100",
				"sv":  "--0vs",
				"sw":  "--0ws",
				"sx":  "--0x1c",
				"t":   "'8'",
				"ta":  "~.8",
				"tas": "%8",
				"u":   "56",
				"ub":  "0b11.1000",
				"uv":  "0v1o",
				"uw":  "0wU",
				"ux":  "0x38",
			},
		},
		{
			dec: "62565",
			exp: map[Aura]string{
				"dr": "~s0..0000.0000.0000.f465",
				"p":  "~bonwet",
				"s":  "-31.283",
				"sb": "-0b111.1010.0011.0011",
				"sv": "-0vuhj",
				"sw": "-0w7EP",
				"sx": "-0x7a33",
				"u":  "62.565",
				"ub": "0b1111.0100.0110.0101",
				"uv": "0v1t35",
				"uw": "0wfhB",
				"ux": "0xf465",
			},
		},
		{
			dec: "1000056",
			exp: map[Aura]string{
				"dr": "~s0..0000.0000.000f.4278",
				"p":  "~dozsun-dapfel",
				"s":  "--500.028",
				"sb": "--0b111.1010.0001.0011.1100",
				"sv": "--0vf89s",
				"sw": "--0w1W4Y",
				"sx": "--0x7.a13c",
				"u":  "1.000.056",
				"ub": "0b1111.0100.0010.0111.1000",
				"uv": "0vugjo",
				"uw": "0w3Q9U",
				"ux": "0xf.4278",
			},
		},
		{
			dec: "100000056",
			exp: map[Aura]string{
				"dr": "~s0..0000.0000.05f5.e138",
				"p":  "~litsen-larbes",
				"r":  "0x5f5e138",
				"rd": ".~4.94065923e-316",
				"rh": ".~~-6.68e2",
				// "rq": ".~~~6.47517875e-4958",
				"rs": ".2.3122422e-35",
				"s":  "--50.000.028",
				"sb": "--0b10.1111.1010.1111.0000.1001.1100",
				"sv": "--0v1.fls4s",
				"sw": "--0w2-L2s",
				"sx": "--0x2fa.f09c",
				"u":  "100.000.056",
				"ub": "0b101.1111.0101.1110.0001.0011.1000",
				"uv": "0v2.vbo9o",
				"uw": "0w5Zu4U",
				"ux": "0x5f5.e138",
			},
		},
		{
			dec: "50000000495056",
			exp: map[Aura]string{
				"dr": "~s0..0000.2d79.8844.add0",
				"p":  "~boltud-nodsub-pocnyd",
				"s":  "--25.000.000.247.528",
				"sb": "--0b1.0110.1011.1100.1100.0100.0010.0010.0101.0110.1110.1000",
				"sv": "--0vmnj2.24ln8",
				"sw": "--0w5HP.48BrE",
				"sx": "--0x16bc.c422.56e8",
				"u":  "50.000.000.495.056",
				"ub": "0b10.1101.0111.1001.1000.1000.0100.0100.1010.1101.1101.0000",
				"uv": "0v1df64.49beg",
				"uw": "0wbnC.8haTg",
				"ux": "0x2d79.8844.add0",
			},
		},
		{
			dec: "324856418037915076923468482958044471810",
			exp: map[Aura]string{
				"dr": "~d203825251263059.h10.m53.s26..4205.e38f.c19c.a202",
				"p":  "~bonwet-dopzod-marnec-litpub--dapper-walrus-digleg-mogbud",
				"s":  "--162.428.209.018.957.538.461.734.241.479.022.235.905",
				"sv": "--0v3.q6a4g.0040g.b9i20.nhovg.csk81",
				"sw": "--0w1W.cEA00.822QO.42Ysv.wPB41",
				"sx": "--0x7a32.8900.0080.82d3.2102.f1c7.e0ce.5101",
				"u":  "324.856.418.037.915.076.923.468.482.958.044.471.810",
				"uv": "0v7.kck90.00810.mj441.f3hv0.pp8g2",
				"uw": "0w3Q.ph800.g45FA.85UU~.1Da82",
				"ux": "0xf465.1200.0101.05a6.4205.e38f.c19c.a202",
			},
		},
		{
			dec: "3180018672171963293882650178620901216809935565260671847783682737515400389475150158146305",
			exp: map[Aura]string{
				"dr": "~d1995244880158166508274203861324441490350683664059201588050179639.h13.m19.s28..d554.9838.9f24.ab01",
				"p":  "~dozsut-socbec-milbyn--difseg-dinnyx-tirnes-sipmer--matrem-docfes-sardul-mirhep--pansyd-dannul-wolsef-dasrem--fopweb-falbes-patpes-bosnec",
				"s":  "-1.590.009.336.085.981.646.941.325.089.310.450.608.404.967.782.630.335.923.891.841.368.757.700.194.737.575.079.073.153",
				"sv": "-0vpie.tl1d7.62til.plqhf.pvvf0.ibvua.c8vbd.71tqm.4fkt3.ro6la.ic3h7.p4lc1",
				"sw": "-0w3cDt.G5FP2.XaKqW.y~f~L.19v~a.ozWSD.3Tlyf.FQuY6.GGj1N.fABm1",
				"sx": "-0x3.3277.6a16.9cc2.ecab.9aea.2fcf.fef0.497f.f298.8fad.a70f.7562.3e9d.1ef0.6aaa.4c1c.4f92.5581",
				"u":  "3.180.018.672.171.963.293.882.650.178.620.901.216.809.935.565.260.671.847.783.682.737.515.400.389.475.150.158.146.305",
				"uv": "0v1j4t.ra2qe.c5r5b.jbl2v.jvuu1.4nvsk.ohumq.e3rlc.8v9q7.ngdal.4o72f.i9ao1",
				"uw": "0w6peX.kbjC5.SlsRR.5-v~u.2i~-k.N7RJe.7KH4v.jEZUd.lkC3y.v9aI1",
				"ux": "0x6.64ee.d42d.3985.d957.35d4.5f9f.fde0.92ff.e531.1f5b.4e1e.eac4.7d3a.3de0.d554.9838.9f24.ab01",
			},
		},
		{
			dec: "125762588864358",
			exp: map[Aura]string{
				"t":   "'foobar'",
				"ta":  "~.foobar",
				"tas": "%foobar",
			},
		},
		{
			dec: "8684515",
			exp: map[Aura]string{
				"t":   "'ツ'",
				"ta":  "~.ツ",
				"tas": "%ツ",
			},
		},
	}
	for _, test := range testCases {
		i, _ := new(big.Int).SetString(test.dec, 10)
		for aura, exp := range test.exp {
			got := Atom{i: i}.Format(aura)
			if got != exp {
				t.Errorf("`@%v`%v\nexp: %v\ngot: %v", aura, test.dec, exp, got)
			}
		}
	}
}

func TestFloat(t *testing.T) {
	testCases := []struct {
		aura Aura
		exp  map[string]string
	}{
		{
			aura: "rq",
			exp: map[string]string{
				"0x0":                                ".~~~0",
				"0x1":                                ".~~~6.475175119438025110924438958227647e-4966",
				"0x3fff0000000000000000000000000000": ".~~~1",
				"0x80000000000000000000000000000000": ".~~~-0",
				"0x7fff0000000000000000000000000000": ".~~~inf",
				"0xffff0000000000000000000000000000": ".~~~-inf",
				"0x7fff8000000000000000000000000000": ".~~~nan",
			},
		},
		{
			aura: "rs",
			exp: map[string]string{
				"0x0":        ".0",
				"0x1":        ".1e-45",
				"0x3f800000": ".1",
				"0x80000000": ".-0",
				"0x7f800000": ".inf",
				"0xff800000": ".-inf",
				"0x7fc00000": ".nan",
			},
		},
		{
			aura: "rh",
			exp: map[string]string{
				"0x0":    ".~~0",
				"0x1":    ".~~5.96e-8",
				"0x3c00": ".~~1",
				"0x8000": ".~~-0",
				"0x7c00": ".~~inf",
				"0xfc00": ".~~-inf",
				"0x7e00": ".~~nan",
			},
		},
	}
	for _, test := range testCases {
		for hex, exp := range test.exp {
			i, _ := new(big.Int).SetString(hex, 0)
			got := Atom{i: i}.Format(test.aura)
			if got != exp {
				t.Errorf("`@%v`%v\nexp: %v\ngot: %v", test.aura, hex, exp, got)
			}
		}
	}
}

func TestDateAbsolute(t *testing.T) {
	testCases := []struct {
		hex, exp string
	}{
		{
			hex: "0x8000000d2da1efc901fa000000000506",
			exp: "~2020.7.7..02.47.37..01fa.0000.0000.0506",
		},
		{
			hex: "0x8000000d2da1efc901fa000000000000",
			exp: "~2020.7.7..02.47.37..01fa",
		},
		{
			hex: "0x7ffffffe570c16800000000000000000",
			exp: "~1.1.1",
		},
	}
	for _, test := range testCases {
		i, _ := new(big.Int).SetString(test.hex, 0)
		got := Atom{i: i}.Format("da")
		if got != test.exp {
			t.Errorf("`@%v`%v\nexp: %v\ngot: %v", "da", test.hex, test.exp, got)
		}
	}
}
