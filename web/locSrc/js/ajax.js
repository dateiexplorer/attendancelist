const imgPlace = document.querySelector('#currentQR');
const dest = document.querySelector('#destination');
const select = document.querySelector('#locSelector');
const content = document.querySelector('#content');
const startSel = document.querySelector('#startSelect');

var currentLocation;

function getQRCode() {
	let req = new XMLHttpRequest();
	let url = "https://localhost:4443/newAccessTk?loc=" + currentLocation;
	req.open("GET", encodeURI(url), true);
	req.onreadystatechange = () =>
	{
		if(req.readyState == 4 && req.status == 200)
		{
			
			if(req.responseText == "")
			{
				console.log("kein resp");
				setTimeout(getQRCode, 10);
			}
			else
			{
				resp = JSON.parse(req.responseText);
				console.log(resp);
				imgPlace.src = "data:image/png;base64," + resp.Qr;
				let exp = new Date(resp.Expires);
				let dif = (exp - new Date().getTime());
				dest.innerHTML = "Zugnangs-QRCode f" + unescape("%FC") +"r " + resp.Location;
				setTimeout(getQRCode, dif);

			}
		}
	}
	req.send(null);
}

function start() {
	currentLocation = select.value;
	imgPlace.style.display = "initial";
	startSel.style.display = "none";
	getQRCode();
}


function getLocations() {
	let req = new XMLHttpRequest();
	let url = "https://localhost:4443/loc";
	req.open("GET", encodeURI(url), true);
	req.onreadystatechange = () =>
	{
		if(req.readyState == 4 && req.status == 200)
		{
			
			if(req.responseText == "")
			{
				console.log("kein resp");
			}
			else
			{
				var opt = document.createElement('option');
   					opt.value = "";
    				opt.innerHTML = "";
    				opt.selected = true;
    				opt.disabled = true;
					select.appendChild(opt);
				resp = JSON.parse(req.responseText);
				for (var i = resp.length - 1; i >= 0; i--) {
					var opt = document.createElement('option');
   					opt.value = resp[i];
    				opt.innerHTML = resp[i];
					select.appendChild(opt);
				}
				//select.innerHTML = "Zugnangs-QRCode f" + unescape("%FC") +"r " + resp.Location;
			}
		}
	}
	req.send(null);
}

