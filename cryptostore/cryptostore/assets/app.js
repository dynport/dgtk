'use strict';

var cryptoStoreApp = angular.module('cryptoStoreApp',
  ['ngRoute', 'cryptoControllers']
);

//var cryptoStoreApp.

cryptoStoreApp.config(['$routeProvider',
  function($routeProvider) {
    $routeProvider.
      when("/login", {
        templateUrl: "login.html",
        controller: 'Login'
      }).
      when('/passwords/new', {
        templateUrl: "password_new.html",
        controller: 'PwdNewCtrl'
      }).
      when('/passwords/:Name', {
        templateUrl: "password.html",
        controller: 'PwdShowCtrl'
      }).
      when('/passwords', {
        templateUrl: "passwords.html",
        controller: 'PwdListCtrl'
      }).
      otherwise({
        redirectTo: "/passwords"
      });
  }
]);
