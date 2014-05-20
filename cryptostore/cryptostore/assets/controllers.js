'use strict';

var cryptoControllers = angular.module('cryptoControllers', []);


// http://blog.brunoscopelliti.com/authentication-to-a-restful-web-service-in-an-angularjs-web-app
//
// Or use a LoginService or SettingsService as described in http://stackoverflow.com/questions/11938380/global-variables-in-angularjs
cryptoControllers.controller("Login", ["$scope", "$http", "$location", function($scope, $http, $location) {
  $scope.User = {}
  $scope.Login = function() {
    console.log("logging in with", $scope.User);
    $location.path("/");
  }
}])

cryptoControllers.controller("PwdNewCtrl", ["$scope", "$http", "$location", function($scope, $http, $location) {
  $scope.password = {}
  $scope.SavePassword = function() {
    console.log("saving password", $scope.password);
    $http.post("/passwords.json", $scope.password);
    $scope.password = {};
    $location.path("/");
  };
}]);


cryptoControllers.controller("PwdShowCtrl", ["$scope", "$http", "$routeParams", function($scope, $http, $routeParams) {
  $scope.refresh = function() {
    $http.get("passwords/" + $routeParams.Name + ".json").success(function(data){
      $scope.password = data;
    });
  }
  $scope.refresh()
}]);

cryptoControllers.controller('PwdListCtrl', ['$scope', '$http', '$routeParams', '$location', function($scope, $http, $routeParams, $location) {
  //$location.path("/login");
  //return;
  $scope.refresh = function() {
    $http.get("passwords.json").success(function(data){
      $scope.passwords = data;
    });
  }
  $scope.refresh()
}]);
