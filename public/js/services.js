'use strict';

/* Services */


// Demonstrate how to register services
// In this case it is a simple value service.
angular.module('myApp.services', []).
  value('version', '0.1')
  .factory('formDataObject', function() {
    return function(data) {
      var fd = new FormData();
      angular.forEach(data, function(value, key) {
        fd.append(key, value);
      });
      return fd;
    };
  });
