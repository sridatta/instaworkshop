'use strict';


// Declare app level module which depends on filters, and services
angular.module('myApp', ['myApp.filters', 'myApp.services', 'myApp.directives', 'myApp.controllers']).
  config(['$routeProvider', function($routeProvider) {
    $routeProvider.when('/', {templateUrl: 'public/partials/list.html', controller: 'ListCtrl'});
    $routeProvider.when('/attach/:image', {templateUrl: 'public/partials/attach.html', controller: 'AttachCtrl'});
    $routeProvider.when('/create', {templateUrl: 'public/partials/create.html', controller: 'CreateCtrl'});
    $routeProvider.otherwise({redirectTo: '/'});
  }]);
