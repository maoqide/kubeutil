function getQueryVariable(variable) {
	let query = window.location.search.substring(1);
	let vars = query.split("&");
	for (let i=0;i<vars.length;i++) {
			let pair = vars[i].split("=");
			if(pair[0] == variable){return pair[1];}
	}
	return(false);
}

function connect(){
	namespace=getQueryVariable("namespace")
	pod=getQueryVariable("pod")
	container_name=getQueryVariable("container")
	tail=getQueryVariable("tail")
	follow=getQueryVariable("follow")
	if (namespace == false) {
		namespace="default"
	}
	// container_name="nginx-2"
	if (namespace == false || pod == false) {
		alert("cannot get pod")
		return
	}
	url = "ws://"+document.location.host+"/ws/"+namespace+"/"+pod+"/"+container_name+"/logs?"
	if (tail != false) {
		url = url+"&tail="+tail
	}
	if (follow != false) {
		url = url+"&follow="+follow
	}

	console.log(url);
	let term = new Terminal({
		// "cursorBlink":true,
	});
	if (window["WebSocket"]) {
		term.open(document.getElementById("terminal"));
		// term.write("logs "+ pod + "...");
		term.toggleFullScreen(true);
		term.fit();
		term.on('data', function (data) {
			conn.send(data)
		});
		conn = new WebSocket(url);
		conn.onopen = function(e) {
		};
		conn.onmessage = function(event) {
			term.writeln(event.data)
			// term.write(event.data)
		};
		conn.onclose = function(event) {
			if (event.wasClean) {
				console.log(`[close] Connection closed cleanly, code=${event.code} reason=${event.reason}`);
			} else {
				console.log('[close] Connection died');
				term.writeln("")
			}
			// term.write('Connection Reset By Peer! Try Refresh.');
		};
		conn.onerror = function(error) {
			console.log('[error] Connection error');
			term.write("error: "+error.message);
			term.destroy();
		};
	} else {
		var item = document.getElementById("terminal");
		item.innerHTML = "<h2>Your browser does not support WebSockets.</h2>";
	}
}
