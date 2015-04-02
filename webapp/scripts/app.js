'use strict';

/**
 * @ngdoc overview
 * @name mgmApp
 * @description
 * # mgmApp
 *
 * Main module of the application.
 */
angular
  .module('mgmApp', [
    'ngAnimate',
    'ngCookies',
    'ngMessages',
    'ngResource',
    'ngRoute',
    'ngSanitize',
    'ngTouch'
  ])
  .config(function ($routeProvider) {
    $routeProvider
    .when('/', {
      templateUrl: 'views/main.html'
    })
    .when('/login', {
      templateUrl: 'views/login.html',
      controller: 'LoginCtrl'
    })
    .when('/about', {
      templateUrl: 'views/about.html'
    })
    .when('/forgotpass', {
      templateUrl: 'views/forgotpass.html',
      controller: 'PasswordresetCtrl'
    })
    .when('/register', {
      templateUrl: 'views/register.html',
      controller: 'RegisterCtrl'
    })
    .otherwise({
      redirectTo: '/'
    });
  });
