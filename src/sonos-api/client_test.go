package sonos

import (
	log "github.com/sirupsen/logrus"
	"testing"
	"time"
)

const accessToken = ""
const refreshToken = ""
const hubToken = ""

func TestClient_GetHousehold(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	client := NewClient("beta",accessToken,refreshToken)
	client.SetHubAuthToken(hubToken)

	household , err := client.GetHousehold()
	if err != nil{
		t.Fatal("Can't retrieve household . Err:",err.Error())
	}
	if len(household)==0 {
		t.Fatal("Empty household")
	}else {
		t.Log(household)
		t.Log("All good")
	}
}

func TestClient_GetGroupsAndPlayers(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	client := NewClient("beta",accessToken,refreshToken)
	client.SetHubAuthToken(hubToken)

	household , err := client.GetHousehold()
	if err != nil{
		t.Fatal("Can't retrieve household . Err:",err.Error())
	}
	if len(household)==0 {
		t.Fatal("Empty household")
	}else {
		t.Log(household)
		t.Log("All good")
	}
	groups,players,err := client.GetGroupsAndPlayers(household[0].ID)
	if err != nil {
		t.Fatal("Can't get groups and Players , Err:",err.Error())
	}
	if len(groups) == 0 || len(players) == 0 {
		t.Fatal("List or groups or players is empty")
	}
	t.Log(groups[0].GroupId)
	t.Log(players[0])
	t.Log("All good")
}

func TestClient_PlaybackSet(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	client := NewClient("beta",accessToken,refreshToken)
	client.SetHubAuthToken(hubToken)

	household , err := client.GetHousehold()
	if err != nil{
		t.Fatal("Can't retrieve household . Err:",err.Error())
	}
	if len(household)==0 {
		t.Fatal("Empty household")
	}else {
		t.Log(household)
		t.Log("All good")
	}

	groups,players,err := client.GetGroupsAndPlayers(household[0].ID)
	if err != nil {
		t.Fatal("Can't get groups and Players , Err:",err.Error())
	}
	if len(groups) == 0 || len(players) == 0 {
		t.Fatal("List or groups or players is empty")
	}

	client.PlaybackSet("toggle_play_pause",groups[0].GroupId)
}

func TestClient_VolumeSet(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	client := NewClient("beta",accessToken,refreshToken)
	client.SetHubAuthToken(hubToken)

	household , err := client.GetHousehold()
	if err != nil{
		t.Fatal("Can't retrieve household . Err:",err.Error())
	}
	if len(household)==0 {
		t.Fatal("Empty household")
	}else {
		t.Log(household)
		t.Log("All good")
	}

	groups,players,err := client.GetGroupsAndPlayers(household[0].ID)
	if err != nil {
		t.Fatal("Can't get groups and Players , Err:",err.Error())
	}
	if len(groups) == 0 || len(players) == 0 {
		t.Fatal("List or groups or players is empty")
	}
	groupId := groups[0].GroupId
	volObj , err := client.VolumeGet(groupId)
	if err != nil {
		t.Fatal("Can't get volume object , Err:",err.Error())
	}
	currVolume := int64(volObj.Volume)
	t.Log("Current volume level = ",currVolume)
	_,err = client.VolumeSet(30,groupId)
	if err != nil {
		t.Fatal("Can't set volume , Err:",err.Error())
	}
	time.Sleep(time.Second*3)
	_,err = client.VolumeSet(currVolume,groupId)
	if err != nil {
		t.Fatal("Can't set volume , Err:",err.Error())
	}

}

func TestClient_AudioClipLoad(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	client := NewClient("beta",accessToken,refreshToken)
	client.SetHubAuthToken(hubToken)

	household , err := client.GetHousehold()
	if err != nil{
		t.Fatal("Can't retrieve household . Err:",err.Error())
	}
	if len(household)==0 {
		t.Fatal("Empty household")
	}else {
		t.Log(household)
		t.Log("All good")
	}

	groups,players,err := client.GetGroupsAndPlayers(household[0].ID)
	if err != nil {
		t.Fatal("Can't get groups and Players , Err:",err.Error())
	}
	if len(groups) == 0 || len(players) == 0 {
		t.Fatal("List or groups or players is empty")
	}

	client.AudioClipLoad("https://storage.googleapis.com/fh-flow-repo/burglar_alarm_voice.mp3",50,players[0].Id)
	//client.AudioClipLoad("http://192.168.86.43/audio/burglar_alarm_voice.mp3",30,players[0].Id)
	//client.AudioClipLoad("http://192.168.86.43/audio/lookdave.wav",50,players[0].Id)
}

