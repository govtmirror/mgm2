'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:AccountCtrl
 * @description
 * # AccountCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('AccountCtrl', function ($scope, $location, $timeout, FileUploader, mgm) {

    if ($scope.auth === undefined || $scope.auth === {}) {
      $location.url('/loading');
    }

    $scope.account = {
      UserID: '',
      Name: '',
      AccessLevel: '',
      Email: ''
    };

    $scope.password = {
      password: '',
      confirm: ''
    };

    $scope.iarName = '';

    for (var uuid in mgm.activeUsers) {
      if (uuid === $scope.auth.UUID) {
        angular.copy(mgm.activeUsers[uuid], $scope.account);
      }
    }

    $scope.$on('UserUpdate', function (event, user) {
      if (user.UserID === $scope.auth.UUID && user !== $scope.auth) {
        $timeout(function () {
          angular.copy(user, $scope.account);
        });
      }
    });

    $scope.setPassword = function () {
      $scope.password.passwordError = '';
      if ($scope.password.password === undefined || $scope.password.password === '') {
        $scope.password.passwordError = 'Password cannot be blank';
        return;
      }
      if ($scope.password.confirm === undefined || $scope.password.confirm === '') {
        $scope.password.passwordError = 'Password confirmation is blank';
        return;
      }
      if ($scope.password.confirm !== $scope.password.password) {
        $scope.password.passwordError = 'Passwords do not match';
        return;
      }

      mgm.ws.emit('SetPassword', angular.toJson({
        UserID: $scope.auth.UUID,
        Password: $scope.password.password
      }), function (response) {
        response = angular.fromJson(response);
        if(response.Success === true){
          alertify.success('Password updated successfully');
          $timeout(function(){
            $scope.password.password = '';
            $scope.password.confirm = '';
          });
        } else {
          alertify.error(response.Message);
        }
      });
    };

    $scope.iar = {
      file: undefined,
      message: '',
      upload: function () {

        alertify.confirm('You are scheduling an IAR upload.  It may not happen immediately, but you do not need to be logged in for it to succeed.', function (e) {
          if (e) {
            mgm.request('IarUpload', {
              UserID: $scope.auth.UUID,
              Filename: $scope.iarName,
            }, function (success, message) {
              $timeout(function () {
                if (success === true) {
                  $scope.uploader.queue.forEach(function (item) {
                    console.log('Uploading Iar file');
                    item.url = 'upload/' + message;
                    item.removeAfterUpload = true;
                    item.upload();
                  });
                } else {
                  alertify.error(message);
                }
              });
            });
          }
        });
      }
    };

    $scope.uploader = new FileUploader({
      url: 'upload',
    });

    $scope.uploader.filters.push({
      name: 'oarFilter',
      fn: function (item, options) {
        var fileExt = item.name.slice(item.name.lastIndexOf('.') + 1);
        return fileExt === 'iar';
      }
    });

    $scope.uploader.onWhenAddingFileFailed = function (item, filter, options) {
      alertify.error('File ' + item.name + ' does not appear to be an iar file');
    };
    $scope.uploader.onAfterAddingAll = function (addedFileItems) {
      $scope.iarName = addedFileItems[0].file.name;
    };

    $scope.uploader.onCompleteAll = function () {
      alertify.success('Iar file uploaded to MGM, load into user account now pending.')
    };

  });
