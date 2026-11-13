// Code generated from OpenLoco's Jukebox.cpp (kMusicInfo), Environment.h /
// Environment.cpp (PathId → sub path) and Config.h (PlaylistItem) by
// tools/gen-jukebox/gen.py; DO NOT EDIT.

package jukebox

import "github.com/LaPingvino/goloco/pkg/i18n"

// musicInfo is kMusicInfo (0x004FE910): each track's file under the
// Locomotion folder, its title and the years of its era; Key is its
// PlaylistItem name, the key of openloco.yml's audio.customJukebox.
var musicInfo = [NumMusicTracks]MusicInfo{
	{File: "20s1.DAT", Key: "chugginAlong", TitleID: i18n.MusicChugginAlong, StartYear: 1925, EndYear: 1933},
	{File: "20s2.DAT", Key: "longDustyRoad", TitleID: i18n.MusicLongDustyRoad, StartYear: 1927, EndYear: 1935},
	{File: "20s4.DAT", Key: "flyingHigh", TitleID: i18n.MusicFlyingHigh, StartYear: 1932, EndYear: 1940},
	{File: "50s1.DAT", Key: "gettinOnTheGas", TitleID: i18n.MusicGettinOnTheGas, StartYear: 1956, EndYear: 1964},
	{File: "50s2.DAT", Key: "jumpinTheRails", TitleID: i18n.MusicJumpinTheRails, StartYear: 1953, EndYear: 1961},
	{File: "70s1.DAT", Key: "smoothRunning", TitleID: i18n.MusicSmoothRunning, StartYear: 1976, EndYear: 1984},
	{File: "70s2.DAT", Key: "trafficJam", TitleID: i18n.MusicTrafficJam, StartYear: 1973, EndYear: 1981},
	{File: "70s3.DAT", Key: "neverStopTilYouGetThere", TitleID: i18n.MusicNeverStopTilYouGetThere, StartYear: 1970, EndYear: 1978},
	{File: "80s1.DAT", Key: "soaringAway", TitleID: i18n.MusicSoaringAway, StartYear: 1990, EndYear: EndOfTime},
	{File: "90s1.DAT", Key: "technoTorture", TitleID: i18n.MusicTechnoTorture, StartYear: 1993, EndYear: EndOfTime},
	{File: "90s2.DAT", Key: "everlastingHighRise", TitleID: i18n.MusicEverlastingHighRise, StartYear: 1996, EndYear: EndOfTime},
	{File: "rag3.DAT", Key: "solace", TitleID: i18n.MusicSolace, StartYear: 1912, EndYear: 1920},
	{File: "Chrysanthemum.DAT", Key: "chrysanthemum", TitleID: i18n.MusicChrysanthemum, StartYear: BeginningOfTime, EndYear: 1911},
	{File: "Eugenia.DAT", Key: "eugenia", TitleID: i18n.MusicEugenia, StartYear: BeginningOfTime, EndYear: 1908},
	{File: "Rag2.DAT", Key: "theRagtimeDance", TitleID: i18n.MusicTheRagtimeDance, StartYear: 1909, EndYear: 1917},
	{File: "Rag1.DAT", Key: "easyWinners", TitleID: i18n.MusicEasyWinners, StartYear: BeginningOfTime, EndYear: 1914},
	{File: "20s3.DAT", Key: "settingOff", TitleID: i18n.MusicSettingOff, StartYear: 1929, EndYear: 1937},
	{File: "40s1.DAT", Key: "aTravellersSerenade", TitleID: i18n.MusicATravellersSerenade, StartYear: 1940, EndYear: 1948},
	{File: "40s2.DAT", Key: "latinoTrip", TitleID: i18n.MusicLatinoTrip, StartYear: 1943, EndYear: 1951},
	{File: "50s3.DAT", Key: "aGoodHeadOfSteam", TitleID: i18n.MusicAGoodHeadOfSteam, StartYear: 1950, EndYear: 1958},
	{File: "40s3.DAT", Key: "hopToTheBop", TitleID: i18n.MusicHopToTheBop, StartYear: 1946, EndYear: 1954},
	{File: "80s2.DAT", Key: "theCityLights", TitleID: i18n.MusicTheCityLights, StartYear: 1980, EndYear: 1988},
	{File: "60s1.DAT", Key: "steaminDownTown", TitleID: i18n.MusicSteaminDownTown, StartYear: 1960, EndYear: 1968},
	{File: "80s3.DAT", Key: "brightExpectations", TitleID: i18n.MusicBrightExpectations, StartYear: 1983, EndYear: 1991},
	{File: "60s2.DAT", Key: "moStation", TitleID: i18n.MusicMoStation, StartYear: 1963, EndYear: 1971},
	{File: "60s3.DAT", Key: "farOut", TitleID: i18n.MusicFarOut, StartYear: 1966, EndYear: 1974},
	{File: "80s4.DAT", Key: "runningOnTime", TitleID: i18n.MusicRunningOnTime, StartYear: 1986, EndYear: 1994},
	{File: "20s5.DAT", Key: "getMeToGladstoneBay", TitleID: i18n.MusicGetMeToGladstoneBay, StartYear: 1918, EndYear: 1926},
	{File: "20s6.DAT", Key: "sandyTrackBlues", TitleID: i18n.MusicSandyTrackBlues, StartYear: 1921, EndYear: 1929},
	{File: "CSS5.DAT", Key: "locomotionTitle", TitleID: i18n.MusicLocomotionTitle, StartYear: EndOfTime, EndYear: BeginningOfTime},
}
