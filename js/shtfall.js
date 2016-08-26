window.onload = function() {
    var conn;

    function start() {
        if (!window["WebSocket"]) {
            return false;
        }

        conn = new WebSocket("ws://localhost:8080/ws");

        conn.onopen = function() {
            console.log("Connection established")
        };

        conn.onclose = function(evt) {
            console.log("Connection refused")

            setTimeout(function(){
                start()
            }, 5000);
        };

        conn.onmessage = function (evt) {
            console.log(evt)
        };
    }

    start();
};
