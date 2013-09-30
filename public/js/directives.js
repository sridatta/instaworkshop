'use strict';

/* Directives */


angular.module('myApp.directives', []).
  directive('appVersion', ['version', function(version) {
    return function(scope, elm, attrs) {
      elm.text(version);
    };
  }])
  .directive('fileChange', function() {
    return function(scope, element, attrs) {
      element[0].onchange = function() {
        scope[attrs['fileChange']](element[0])
      }
    }
  })
  .directive('terminal', function(){
    return function(scope, elm, attrs) {
      var image = scope.workshopImage
      var socket = new WebSocket("ws://127.0.0.1:9000/api/workshops/"+scope.workshopImage)
      var term = new Terminal({
        cols: 80,
        rows: 24,
        screenKeys: true
      });

      term.on('data', function(data) {
        socket.send(data);
      });

      term.on('title', function(title) {
        document.title = title;
      });

      term.open(document.body);

      term.write('\x1b[31mWelcome to term.js!\x1b[m\r\n');

      socket.onmessage = function(msg) {
        term.write(msg.data);
      };

      //socket.on('disconnect', function() {
        //term.destroy();
      //});
    }
  })


