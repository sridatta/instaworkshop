'use strict';

/* Controllers */

angular.module('myApp.controllers', []).
  controller('ListCtrl', function($scope, $http) {
    $scope.workshops = []
    $http.get("/api/workshops").success(function(workshops){
      $scope.workshops = workshops
    })
  })
  .controller('AttachCtrl', function($scope, $routeParams) {
    $scope.workshopImage = $routeParams.image;
  })
  .controller("CreateCtrl", function($scope, $http, formDataObject, $location) {
    $scope.createImage = function(){
      if(!$scope.script) {
        return
      }
      $http({
        method: 'POST',
        url: '/api/workshops',
        headers: {
          'Content-Type': false
        },
        data : {
          file: $scope.theFile,
          name: $scope.name,
          script: $scope.script
        },
        transformRequest: formDataObject
      }).success(function() {
        $location.path("/")
      })
    }

    $scope.setFile = function(element){
      $scope.$apply(function() {
        $scope.theFile = element.files[0];
      })
    }
  })
