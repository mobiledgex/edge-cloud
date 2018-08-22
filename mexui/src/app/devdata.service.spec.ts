import { TestBed, inject } from '@angular/core/testing';

import { DevdataService } from './devdata.service';

describe('DevdataService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [DevdataService]
    });
  });

  it('should be created', inject([DevdataService], (service: DevdataService) => {
    expect(service).toBeTruthy();
  }));
});
