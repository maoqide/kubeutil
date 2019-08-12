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
	container_name=getQueryVariable("container_name")
	if (namespace == false || pod == false || container_name == false) {
		alert("无法获取到容器，请联系管理员")
		return
	}
	console.log(namespace ,pod ,container_name)
	url = "ws://"+document.location.host+"/ws/"+namespace+"/"+pod+"/"+container_name+"/webshell"
	console.log(url);
	let term = new Terminal({
		"cursorBlink":true,
	});
	if (window["WebSocket"]) {
		term.open(document.getElementById("terminal"));
		term.write("connecting to pod "+ pod + "...")
		term.fit();
		// term.toggleFullScreen(true);
		term.on('data', function (data) {
			msg = {operation: "stdin", data: data}
			conn.send(JSON.stringify(msg))
		});
		term.on('resize', function (size) {
			console.log("resize: " + size)
			msg = {operation: "resize", cols: size.cols, rows: rows}
			conn.send(JSON.stringify(msg))
		});

		conn = new WebSocket(url);
		conn.onopen = function(e) {
			term.write("\r");
			msg = {operation: "stdin", data: "export TERM=xterm && clear \r"}
			conn.send(JSON.stringify(msg))
			// term.clear()
		};
		conn.onmessage = function(event) {
			msg = JSON.parse(event.data)
			if (msg.operation === "stdout") {
				term.write(msg.data)
			} else {
				console.log("invalid msg operation: "+msg)
			}
		};
		conn.onclose = function(event) {
			if (event.wasClean) {
				console.log(`[close] Connection closed cleanly, code=${event.code} reason=${event.reason}`);
			} else {
				console.log('[close] Connection died');
				term.writeln("")
			}
			term.write('Connection Reset By Peer! Try Refresh.');
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