window.document.onload = getQRCode();
const imgPlace = document.querySelector('#currentQR');

//Function to get the Accestokeninformation
function getQRCode() {
	let req = new XMLHttpRequest();
	let url = "https://localhost:4443/newAccessTk?loc=DHBW Mosbach";
	req.open("GET", encodeURI(url), true);
	req.onreadystatechange = () =>
	{
		if(req.readyState == 4 && req.status == 200)
		{
			resp = JSON.parse(req.responseText);
			console.log(resp);
			if(resp == "")
			{
				console.log("kein resp");
				setTimeout(getQRCode, 1000);
			}
			else
			{
				console.log('drin');
				imgPlace.src = "data:image/png;base64," + resp.Qr;
				let exp = new Date(resp.Expires);
				let dif = (exp - new Date().getTime());
				setTimeout(getQRCode, dif);

			}
		}
	}
	req.send(null);
}