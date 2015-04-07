'use strict';

describe('Service: mgm', function () {

  // load the service's module
  beforeEach(module('mgmApp'));

  // instantiate service
  var mgm;
  beforeEach(inject(function (_mgm_) {
    mgm = _mgm_;
  }));

  it('should do something', function () {
    expect(!!mgm).toBe(true);
  });

});
