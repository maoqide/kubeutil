function geturi() {
　　　　var url = document.location.toString();
　　　　var arrUrl = url.split("//");
　　　　var start = arrUrl[1].indexOf("/");
　　　　var uri = arrUrl[1].substring(start+1);
　　　　if(uri.indexOf("?") != -1){
　　　　　　uri = uri.split("?")[0];
　　　　}
　　　　return uri;
}

function getParams(uri) {
	var params = uri.split("/")
	if (params.length != 4 ){
		return []
	}
	namespace=params[1]
	pod=params[2]
	container_name=params[3]
	return[namespace, pod, container_name]
}

function connect(params){
	if (params.length == 0) {
		alert("无法获取到容器，请联系管理员")
		return
	}
	namespace=params[0]
	pod=params[1]
	container_name=params[2]
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
			}
			term.write('Connection Reset By Peer');
		};
		conn.onerror = function(error) {
			term.write("error: "+error.message);
			term.destroy();
		};
	} else {
		var item = document.getElementById("terminal");
		item.innerHTML = "<h2>Your browser does not support WebSockets.</h2>";
	}
}