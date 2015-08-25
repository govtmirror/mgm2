'use strict';

describe('Directive: mgmConsole', function () {

  // load the directive's module
  beforeEach(module('mgmApp'));

  var element,
    scope;

  beforeEach(inject(function ($rootScope) {
    scope = $rootScope.$new();
  }));

  it('should make hidden element visible', inject(function ($compile) {
    element = angular.element('<mgm-console></mgm-console>');
    element = $compile(element)(scope);
    expect(element.text()).toBe('this is the mgmConsole directive');
  }));
});
