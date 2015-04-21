'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:MapCtrl
 * @description
 * # MapCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('MapCtrl', function ($scope, $location) {

    if ($scope.auth === undefined) {
      $location.url("/");
    }
  });
