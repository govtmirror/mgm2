'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:AccountCtrl
 * @description
 * # AccountCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('AccountCtrl', function ($scope, mgm) {
    $scope.account = mgm.account;

    $scope.$on("AccountChange", function (event, data) {
      $scope.account = mgm.account;
      $scope.$apply();
    });

  });
