// This source file is part of the attendance list project
// as a part of the go lecture by H. Neemann.
// For this reason you have no permission to use, modify or
// share this code without the agreement of the authors.
//
// Matriculation numbers of the authors: 5703004, 5736465

// Divs for success and error
const divSuccess = document.getElementById("success");
const divError = document.getElementById("error");

// Elements to manipulate
const qr = document.getElementById("qr");
const loc = document.getElementById("loc");
const countdown = document.getElementById("countdown");

// The sleep function returns a promise for a async await syntax.
// The code stops and continues after an amount of time.
const sleep = (milliseconds) => {
	return new Promise(resolve => setTimeout(resolve, milliseconds))
}

async function getQRCode() {
	"use strict";

	let url = `https://${window.location.host}/api/tokens?location=${loc.innerHTML}`;

	let error = 0;
	while (error < 1) {
		try {
			let response = await fetch(url);

			if (response.ok) {
				let json = await response.json();
		
				qrcode.src = "data:image/png;base64," + json.qr;
		
				let currentUnixTime = Date.now();
				let diff = json.exp * 1000 - currentUnixTime;
				
				// If the access code already is expired wait an amount of time
				// for the next fetch.
				// This happens if the fetch started before the server has
				// generated the new access token.
				if (diff < 0) {
					await sleep(500);
				}
		
				await sleep(diff);
			}
		} catch {
			error++;
		}
	}

	// Try reload page if something went wrong.
	divSuccess.hidden = true;
	divError.hidden = false;

	let secondsUntilRetry = 10;
	let t = setInterval(() => {
		countdown.innerHTML = `Retry in <strong>${secondsUntilRetry}</strong> second(s).`
		secondsUntilRetry--;

		if (secondsUntilRetry == 0) {
			clearInterval(t);
			window.location.reload();
		}
	}, 1000);
}

// Execute function on load
window.onload = getQRCode;
