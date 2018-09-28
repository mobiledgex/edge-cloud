# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
import grpc

import flavor_pb2 as flavor__pb2
import result_pb2 as result__pb2


class FlavorApiStub(object):
  # missing associated documentation comment in .proto file
  pass

  def __init__(self, channel):
    """Constructor.

    Args:
      channel: A grpc.Channel.
    """
    self.CreateFlavor = channel.unary_unary(
        '/edgeproto.FlavorApi/CreateFlavor',
        request_serializer=flavor__pb2.Flavor.SerializeToString,
        response_deserializer=result__pb2.Result.FromString,
        )
    self.DeleteFlavor = channel.unary_unary(
        '/edgeproto.FlavorApi/DeleteFlavor',
        request_serializer=flavor__pb2.Flavor.SerializeToString,
        response_deserializer=result__pb2.Result.FromString,
        )
    self.UpdateFlavor = channel.unary_unary(
        '/edgeproto.FlavorApi/UpdateFlavor',
        request_serializer=flavor__pb2.Flavor.SerializeToString,
        response_deserializer=result__pb2.Result.FromString,
        )
    self.ShowFlavor = channel.unary_stream(
        '/edgeproto.FlavorApi/ShowFlavor',
        request_serializer=flavor__pb2.Flavor.SerializeToString,
        response_deserializer=flavor__pb2.Flavor.FromString,
        )


class FlavorApiServicer(object):
  # missing associated documentation comment in .proto file
  pass

  def CreateFlavor(self, request, context):
    """Create a Flavor
    """
    context.set_code(grpc.StatusCode.UNIMPLEMENTED)
    context.set_details('Method not implemented!')
    raise NotImplementedError('Method not implemented!')

  def DeleteFlavor(self, request, context):
    """Delete a Flavor
    """
    context.set_code(grpc.StatusCode.UNIMPLEMENTED)
    context.set_details('Method not implemented!')
    raise NotImplementedError('Method not implemented!')

  def UpdateFlavor(self, request, context):
    """Update a Flavor
    """
    context.set_code(grpc.StatusCode.UNIMPLEMENTED)
    context.set_details('Method not implemented!')
    raise NotImplementedError('Method not implemented!')

  def ShowFlavor(self, request, context):
    """Show Flavors
    """
    context.set_code(grpc.StatusCode.UNIMPLEMENTED)
    context.set_details('Method not implemented!')
    raise NotImplementedError('Method not implemented!')


def add_FlavorApiServicer_to_server(servicer, server):
  rpc_method_handlers = {
      'CreateFlavor': grpc.unary_unary_rpc_method_handler(
          servicer.CreateFlavor,
          request_deserializer=flavor__pb2.Flavor.FromString,
          response_serializer=result__pb2.Result.SerializeToString,
      ),
      'DeleteFlavor': grpc.unary_unary_rpc_method_handler(
          servicer.DeleteFlavor,
          request_deserializer=flavor__pb2.Flavor.FromString,
          response_serializer=result__pb2.Result.SerializeToString,
      ),
      'UpdateFlavor': grpc.unary_unary_rpc_method_handler(
          servicer.UpdateFlavor,
          request_deserializer=flavor__pb2.Flavor.FromString,
          response_serializer=result__pb2.Result.SerializeToString,
      ),
      'ShowFlavor': grpc.unary_stream_rpc_method_handler(
          servicer.ShowFlavor,
          request_deserializer=flavor__pb2.Flavor.FromString,
          response_serializer=flavor__pb2.Flavor.SerializeToString,
      ),
  }
  generic_handler = grpc.method_handlers_generic_handler(
      'edgeproto.FlavorApi', rpc_method_handlers)
  server.add_generic_rpc_handlers((generic_handler,))
