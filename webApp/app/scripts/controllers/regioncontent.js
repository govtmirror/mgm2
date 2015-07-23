'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:RegioncontentCtrl
 * @description
 * # RegioncontentCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('RegioncontentCtrl', function ($scope, $modalInstance, FileUploader, region) {
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

    $scope.uploadOar = function () {
      alertify.confirm('Are you sure you wish to upload this oar file?  It will overwrite all content currently in the region.', function(e){
        if(e){
          $scope.uploader.queue.forEach(function(item){
            console.log(item);
            item.url = 'upload/12345';
            item.removeAfterUpload = true;
            item.upload();
          });
        }
      })
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
      console.log(addedFileItems[0]);
    };

    $scope.uploader.onBeforeUploadItem = function (item) {
      console.info('onBeforeUploadItem', item);
    };
    $scope.uploader.onProgressItem = function (fileItem, progress) {
      console.info('onProgressItem', fileItem, progress);
    };
    $scope.uploader.onProgressAll = function (progress) {
      console.info('onProgressAll', progress);
    };
    $scope.uploader.onSuccessItem = function (fileItem, response, status, headers) {
      console.info('onSuccessItem', fileItem, response, status, headers);
    };
    $scope.uploader.onErrorItem = function (fileItem, response, status, headers) {
      console.info('onErrorItem', fileItem, response, status, headers);
    };
    $scope.uploader.onCancelItem = function (fileItem, response, status, headers) {
      console.info('onCancelItem', fileItem, response, status, headers);
    };
    $scope.uploader.onCompleteItem = function (fileItem, response, status, headers) {
      console.info('onCompleteItem', fileItem, response, status, headers);
    };
    $scope.uploader.onCompleteAll = function () {
      console.info('onCompleteAll');
    };

    //console.info('uploader', $scope.uploader);
  });
