'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:NavCtrl
 * @description
 * # NavCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp').controller('NavCtrl', function ($scope, $location, $timeout, mgmPublic, mgm) {

  $scope.loggedIn = false;
  $scope.$on('AuthChange', function (event, data) {
    console.log('auth changed to ' + data);
    $scope.loggedIn = data;
    if ($scope.loggedIn) {
      mgm.connect();
      $timeout(function () {
        $location.url('/loading');
      }, 0.1);
    } else {
      $timeout(function () {
        $location.url('/');
      }, 0.1);
    }
  });

  $scope.isActive = function (route) {
    return $location.path().indexOf(route) === 0;
  };

  $scope.logout = function () {
    mgm.disconnect();
    mgmPublic.logout();
  };

  $scope.task = {
    showBar: false,
    toggle: function () {
      $scope.task.showBar = !$scope.task.showBar;
    },
    jobs: {},
    delete: function(job){
      mgm.deleteJob(job).then(
        function(){
          //success
        },
        function(msg){
          console.log(msg);
        }
      );
    }
  };

  $scope.task.jobs = mgm.jobs;
  $scope.$on('JobUpdate', function (event, job) {
      $timeout(function(){
        $scope.task.jobs[job.ID] = job;
      })
  });
  $scope.$on('JobDeleted', function (event, job){
    $timeout(function(){
      delete $scope.task.jobs[job.ID];
      alertify.log('Job deleted');
    });
  });
});
