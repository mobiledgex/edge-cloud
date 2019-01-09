﻿using System;

namespace DistributedMatchEngine
{
  [Serializable]
  public class Loc
  {
    public double latitude;
    public double longitude;
    public double horizontal_accuracy;
    public double vertical_accuracy;
    public double altitude;
    public double course;
    public double speed;
    public Timestamp timestamp;
  }
}
