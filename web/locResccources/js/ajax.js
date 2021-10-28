const imgPlace = document.querySelector('#currentQR');
const dest = document.querySelector('#destination');

window.document.onload = getQRCode();


//Function to get the Accestokeninformation
function getQRCode() {
	let req = new XMLHttpRequest();
	let url = "https://localhost:4443/newAccessTk?loc=DHBW MOSBACH";
	req.open("GET", encodeURI(url), true);
	req.onreadystatechange = () =>
	{
		if(req.readyState == 4 && req.status == 200)
		{
			
			if(req.responseText == "")
			{
				console.log("kein resp");
				setTimeout(getQRCode, 1000);
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