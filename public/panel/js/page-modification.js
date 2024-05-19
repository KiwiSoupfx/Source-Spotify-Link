let clientId = "";
let repeatTrackTimeoutId = 0;
let timeout = 0;
//console.log(new URLSearchParams(window.location.search))
//console.log(window.location.search)
console.log(clientId)
let isConnected = false;
let hitRepeat = false;

if (clientId = "") {
    //Update *something* on the page to tell user they didn't set clientid or their browser is weird
}

/*
Any JS devs wanna make repeat getting tracks better ;)
*/

const timer = ms => new Promise(res => setTimeout(res, ms));

document.addEventListener('DOMContentLoaded', function() {
    clientId = new URLSearchParams(window.location.search).get('client_id');
    handleGetTrackButton();
    if (document.getElementById("currTrack").textContent == "No song detected.") {isConnected = false;} else {isConnected = true;} //hopefully useful later
})

function changeTimeout(func) {
    clearTimeout(repeatTrackTimeoutId);
    repeatTrackTimeoutId = setTimeout(func, timeout+1000);
}

function handleGetAuthButton() {
    window.location.href = "https://accounts.spotify.com/en/authorize?client_id="+ clientId + "&redirect_uri=http%3A%2F%2Flocalhost%3A8080&response_type=code&scope=user-read-currently-playing";
}


async function getTrackData() {
    if (!hitRepeat && isConnected) {fetch('/repeatcheck', {method: "GET"}).then(); hitRepeat = true;} //Make sure we're still getting the data if we close the tab
    let trackDataResp = await fetch('/gettrackdata', {
        method: "GET"
    })
    

    let trackData = await trackDataResp.json();
    

    //Update text on page
    if (trackData.track_name != "" && trackData.artists != "") {
        document.getElementById("currTrack").textContent = trackData.track_name + " - " + trackData.artists + " " + trackData.time_left + " left";
        document.getElementById("connectStatus").textContent = "Connected.";
    } else {
        document.getElementById("currTrack").textContent = "No song detected.";
    }
}

//Async so we make sure we get data before updating page
async function handleGetTrackButton() {

    await getTrackData()
    console.log("got track data")

    if (repeatTrackTimeoutId != 0 && timeout != 0) {
        setTrackTimeout().then(() => {changeTimeout(getTrackData);});
    }
}

function handleReverse(str) {
    if (str.length > 1) {
        return str.split("").reverse().join("");
    }
    if (str.length == 0) {return 0;} //Return 0 so we don't error out the ParseInt
    return str;
}

async function setTrackTimeout() {
        //We don't care about anything this spits out except for timeLeft
        let trackDataResp = await fetch('/gettrackdata', {
            method: "GET"
        })

        let trackDataJson = await trackDataResp.json();
        console.log(trackDataJson)
        let timeStr = trackDataJson.time_left;
        //TODO: work out format so we can set up a loop to update the panel
        //hope there's a better way to do this
        //1h59m53.724s
        //00m -> minutes
        //00.000 .> s
        //.000 -> ms or s+1
        timeout = 0;
        let hours = "";
        let minutes = "";
        let seconds = "";
        let milliseconds = "";
        let mode = "";
    
        for (let i = 0; i < timeStr.length; i++) {
            //to go backwards we do .length-i
            let currChar = timeStr[timeStr.length-i-1]
            if (currChar == "s") { //Activate ms
                mode = "ms";
                continue;
            }
            if (currChar == ".") {
                mode = "s";
                continue;
            }
            if (currChar == "m") {
                mode = "m";
                continue;
            }
            if (currChar == "h") {
                mode = "h";
                continue;
            }
            
            if (mode == "ms") {
                milliseconds += currChar
            }
            if (mode == "s") {
                seconds += currChar
            }
            if (mode == "m") {
                minutes += currChar
            }
            if (mode == "h") {
                hours += currChar
            }
        }
        timeout = parseInt(handleReverse(milliseconds)) + (parseInt(handleReverse(seconds)) * 1000) + (parseInt(handleReverse(minutes)) * 60000) + (parseInt(handleReverse(hours)) * 3600000000);

        if (timeout == 0) {timeout = 2000;}
}

async function startRepeat() {
    setTrackTimeout().then(() => {
        changeTimeout(getTrackData)
    })
    while (true) {
        await timer(timeout);
        if (timeout < 1) {timeout = 1000;}
        setTrackTimeout().then(() => {
            changeTimeout(getTrackData)
        })
    }
}

let authButton = document.getElementById("spotGetAuth");

let getTrackButton = document.getElementById("spotGetTrack");

let repeatGetTrackButton = document.getElementById("repeatTrackCheck");

authButton.addEventListener('click', handleGetAuthButton);
getTrackButton.addEventListener('click', () => handleGetTrackButton(), false);
repeatGetTrackButton.addEventListener('click', () => startRepeat(), false);