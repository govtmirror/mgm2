'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:LoadingCtrl
 * @description
 * # LoadingCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('LoadingCtrl', function ($scope, $location, $timeout, mgm) {

    if ($scope.auth === undefined) {
      $location.url("/login");
    }

    $scope.$on("SyncComplete", function (event, data) {
      var url = mgm.popLocation();
      if (url === undefined || url == "/loading" || url == "/login") {
        console.log("no url found, redirecting to account page");
        $timeout(function () {
          $location.url("/account");
        }, 1);
      } else {
        console.log("regirecting to " + url);
        $timeout(function () {
          $location.url(url);
        }, 1);
      }
    });

  });