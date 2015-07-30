'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:RegioncontentCtrl
 * @description
 * # RegioncontentCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('RegioncontentCtrl', function ($scope, $modalInstance, $timeout, FileUploader, mgm, region) {
    $scope.region = region;

    $scope.close = function () {
      $modalInstance.close();
    };

    $scope.saveOar = function(){
      alertify.error('save oar is not implemented');
    }

    $scope.eraseContent = function(){
      alertify.error('erase content is not implemented');
    }

    $scope.uploadOar = function (x, y, merge) {
      alertify.confirm('Are you sure you wish to upload this oar file?  It will overwrite all content currently in the region.', function(e){
        if(e){
          mgm.request('OarUpload', {
            RegionUUID: region.UUID,
            X: x,
            Y: y,
            Merge: merge,
          }, function (success, message) {
            $timeout(function () {
              if (success === true) {
                $scope.uploader.queue.forEach(function (item) {
                  console.log(item);
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

    $scope.onFileSelect = function ($files) {
      console.log($files);
    }

    $scope.oar = {
      uploadFilePresent: false,
      filename: ''
    };

    $scope.uploader = new FileUploader({
      url: 'upload',
    });

    $scope.uploader.filters.push({
      name: 'oarFilter',
      fn: function (item, options) {
        var fileExt = item.name.slice(item.name.lastIndexOf('.')+1);
        return fileExt === 'oar';
      }
    });

    $scope.uploader.onWhenAddingFileFailed = function (item, filter, options) {
      alertify.error('File ' + item.name + ' does not appear to be an oar file');
    };
    $scope.uploader.onAfterAddingAll = function (addedFileItems) {
      $scope.oar.uploadFilePresent = true;
      $scope.oar.filename = addedFileItems[0].file.name;
    };

    $scope.uploader.onCompleteAll = function () {
      alertify.success('Oar file uploaded to MGM, load into region now pending.')
    };
  });
