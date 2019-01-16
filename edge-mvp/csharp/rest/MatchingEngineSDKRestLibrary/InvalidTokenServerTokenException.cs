using System;
namespace DistributedMatchEngine
{
  public class InvalidTokenServerTokenException : Exception
  {
    public InvalidTokenServerTokenException()
    {
    }

    public InvalidTokenServerTokenException(string message)
        : base(message)
    {
    }

    public InvalidTokenServerTokenException(string message, Exception inner)
        : base(message, inner)
    {
    }
  }
}
