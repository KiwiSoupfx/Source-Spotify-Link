let clientId = "";
let repeatTrackTimeoutId = 0;
let timeout = 1000;

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
    getTrackData().then();
})

function updateCheckTimeout() {
    clearTimeout(repeatTrackTimeoutId);
    if (timeout == 0) {timeout = 500;}
    repeatTrackTimeoutId = setTimeout(getTrackData, timeout+100);
}

function handleGetAuthButton() {
    window.location.href = "https://accounts.spotify.com/en/authorize?client_id="+ clientId + "&redirect_uri=http%3A%2F%2Flocalhost%3A8080&response_type=code&scope=user-read-currently-playing";
}

function parseTimeStamp(timeStr) {
    if (!timeStr) {return 500;}
        //TODO: work out format so we can set up a loop to update the panel
        //hope there's a better way to do this
        //1h59m53.724s
        //00m -> minutes
        //00.000 .> s
        //.000 -> ms or s+1
        //???ms -> ms
        let trackTimeout = 0; //Oops. Probably don't set our whole timeout to 0 even if for a second
        let hours = "";
        let minutes = "";
        let seconds = "";
        let milliseconds = "";
        let mode = "";
    
        if (timeStr.includes("ms")) { //It's important that we don't misunderstand milliseconds as minutes.
            milliseconds = parseInt(timeStr.replace("ms", "")); //A little sloppy
            trackTimeout = milliseconds+100;
        } else {
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
            trackTimeout = parseInt(handleReverse(milliseconds)) + (parseInt(handleReverse(seconds)) * 1000) + (parseInt(handleReverse(minutes)) * 60000) + (parseInt(handleReverse(hours)) * 3600000000);
        }
        if (trackTimeout == 0) {trackTimeout = 500;}
        return trackTimeout;
}


async function getTrackData() {
    let trackDataResp = await fetch('/gettrackdata', {method: "GET"});
    

    let trackData = await trackDataResp.json();
    console.log(trackData);

    //Update text on page
    if (trackData.track_name != "" && trackData.artists != "") {
        document.getElementById("currTrack").textContent = trackData.track_name + " - " + trackData.artists + " " + trackData.time_left + " left";
        document.getElementById("connectStatus").textContent = "Connected.";
        isConnected = true;
    } else {
        document.getElementById("currTrack").textContent = "No song detected.";
        isConnected = false; //Not necessarily true
    }
    timeout = parseTimeStamp(trackData.time_left);
    if (!hitRepeat) {fetch('/repeatcheck', {method: "GET"}).then(); startRepeat(); hitRepeat = true; return;} //Make sure we're still getting the data if we close the tab
    updateCheckTimeout();
}

//Async so we make sure we get data before updating page
async function handleGetTrackButton() {
    getTrackData().then(() => {updateCheckTimeout();});
}

function handleReverse(str) {
    if (str.length > 1) {
        return str.split("").reverse().join("");
    }
    if (str.length == 0) {return 0;} //Return 0 so we don't error out the ParseInt
    return str;
}


async function startRepeat() { //Because of how this works, they'll both be evaluated before they get a timeout.
    if (hitRepeat) {return;} //We only need once instance of this running
    if (!isConnected) {return;} //No point in looping if we're not connected.
    for (;;) {
        if (timeout < 1000) {timeout += 1000;}
        await timer(timeout);
        await getTrackData("loop"); //Not a fan of how it runs twice as it evaluates the function because updateCheckTimeout is at the end of this function.
    }
}

let authButton = document.getElementById("spotGetAuth");

let getTrackButton = document.getElementById("spotGetTrack");

let repeatGetTrackButton = document.getElementById("repeatTrackCheck");

authButton.addEventListener('click', handleGetAuthButton);
getTrackButton.addEventListener('click', () => handleGetTrackButton(), false);