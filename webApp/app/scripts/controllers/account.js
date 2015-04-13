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
    $scope.account = {
      UserID: "",
      Name: "",
      Email: "",
      AccessLevel: 0
    };

    $scope.$on("AccountChange", function (event, data) {
      $scope.account = mgm.account;
      $scope.$apply();
    });

  });
