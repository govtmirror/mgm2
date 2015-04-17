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
      .when('/account', {
        templateUrl: 'views/account.html',
        controller: 'AccountCtrl'
      })
      .when('/regions', {
        templateUrl: 'views/regions.html',
        controller: 'RegionsCtrl'
      })
      .when('/grid/:section?', {
        templateUrl: 'views/grid.html',
        controller: 'GridCtrl'
      })
      .when('/map', {
        templateUrl: 'views/map.html',
        controller: 'MapCtrl'
      })
      .when('/users/:section?', {
        templateUrl: 'views/users.html',
        controller: 'UsersCtrl'
      })
      .when('/config/:region?', {
        templateUrl: 'views/config.html',
        controller: 'ConfigCtrl'
      })
      .otherwise({
        redirectTo: '/'
      });
  });
