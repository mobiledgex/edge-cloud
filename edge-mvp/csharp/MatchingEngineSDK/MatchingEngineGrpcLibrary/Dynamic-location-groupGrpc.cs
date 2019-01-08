// <auto-generated>
//     Generated by the protocol buffer compiler.  DO NOT EDIT!
//     source: dynamic-location-group.proto
// </auto-generated>
// Original file comments:
// Dynamic Location Group APIs
//
#pragma warning disable 0414, 1591
#region Designer generated code

using grpc = global::Grpc.Core;

namespace DistributedMatchEngine {
  public static partial class DynamicLocGroupApi
  {
    static readonly string __ServiceName = "distributed_match_engine.DynamicLocGroupApi";

    static readonly grpc::Marshaller<global::DistributedMatchEngine.DlgMessage> __Marshaller_distributed_match_engine_DlgMessage = grpc::Marshallers.Create((arg) => global::Google.Protobuf.MessageExtensions.ToByteArray(arg), global::DistributedMatchEngine.DlgMessage.Parser.ParseFrom);
    static readonly grpc::Marshaller<global::DistributedMatchEngine.DlgReply> __Marshaller_distributed_match_engine_DlgReply = grpc::Marshallers.Create((arg) => global::Google.Protobuf.MessageExtensions.ToByteArray(arg), global::DistributedMatchEngine.DlgReply.Parser.ParseFrom);

    static readonly grpc::Method<global::DistributedMatchEngine.DlgMessage, global::DistributedMatchEngine.DlgReply> __Method_SendToGroup = new grpc::Method<global::DistributedMatchEngine.DlgMessage, global::DistributedMatchEngine.DlgReply>(
        grpc::MethodType.Unary,
        __ServiceName,
        "SendToGroup",
        __Marshaller_distributed_match_engine_DlgMessage,
        __Marshaller_distributed_match_engine_DlgReply);

    /// <summary>Service descriptor</summary>
    public static global::Google.Protobuf.Reflection.ServiceDescriptor Descriptor
    {
      get { return global::DistributedMatchEngine.DynamicLocationGroupReflection.Descriptor.Services[0]; }
    }

    /// <summary>Base class for server-side implementations of DynamicLocGroupApi</summary>
    public abstract partial class DynamicLocGroupApiBase
    {
      public virtual global::System.Threading.Tasks.Task<global::DistributedMatchEngine.DlgReply> SendToGroup(global::DistributedMatchEngine.DlgMessage request, grpc::ServerCallContext context)
      {
        throw new grpc::RpcException(new grpc::Status(grpc::StatusCode.Unimplemented, ""));
      }

    }

    /// <summary>Client for DynamicLocGroupApi</summary>
    public partial class DynamicLocGroupApiClient : grpc::ClientBase<DynamicLocGroupApiClient>
    {
      /// <summary>Creates a new client for DynamicLocGroupApi</summary>
      /// <param name="channel">The channel to use to make remote calls.</param>
      public DynamicLocGroupApiClient(grpc::Channel channel) : base(channel)
      {
      }
      /// <summary>Creates a new client for DynamicLocGroupApi that uses a custom <c>CallInvoker</c>.</summary>
      /// <param name="callInvoker">The callInvoker to use to make remote calls.</param>
      public DynamicLocGroupApiClient(grpc::CallInvoker callInvoker) : base(callInvoker)
      {
      }
      /// <summary>Protected parameterless constructor to allow creation of test doubles.</summary>
      protected DynamicLocGroupApiClient() : base()
      {
      }
      /// <summary>Protected constructor to allow creation of configured clients.</summary>
      /// <param name="configuration">The client configuration.</param>
      protected DynamicLocGroupApiClient(ClientBaseConfiguration configuration) : base(configuration)
      {
      }

      public virtual global::DistributedMatchEngine.DlgReply SendToGroup(global::DistributedMatchEngine.DlgMessage request, grpc::Metadata headers = null, global::System.DateTime? deadline = null, global::System.Threading.CancellationToken cancellationToken = default(global::System.Threading.CancellationToken))
      {
        return SendToGroup(request, new grpc::CallOptions(headers, deadline, cancellationToken));
      }
      public virtual global::DistributedMatchEngine.DlgReply SendToGroup(global::DistributedMatchEngine.DlgMessage request, grpc::CallOptions options)
      {
        return CallInvoker.BlockingUnaryCall(__Method_SendToGroup, null, options, request);
      }
      public virtual grpc::AsyncUnaryCall<global::DistributedMatchEngine.DlgReply> SendToGroupAsync(global::DistributedMatchEngine.DlgMessage request, grpc::Metadata headers = null, global::System.DateTime? deadline = null, global::System.Threading.CancellationToken cancellationToken = default(global::System.Threading.CancellationToken))
      {
        return SendToGroupAsync(request, new grpc::CallOptions(headers, deadline, cancellationToken));
      }
      public virtual grpc::AsyncUnaryCall<global::DistributedMatchEngine.DlgReply> SendToGroupAsync(global::DistributedMatchEngine.DlgMessage request, grpc::CallOptions options)
      {
        return CallInvoker.AsyncUnaryCall(__Method_SendToGroup, null, options, request);
      }
      /// <summary>Creates a new instance of client from given <c>ClientBaseConfiguration</c>.</summary>
      protected override DynamicLocGroupApiClient NewInstance(ClientBaseConfiguration configuration)
      {
        return new DynamicLocGroupApiClient(configuration);
      }
    }

    /// <summary>Creates service definition that can be registered with a server</summary>
    /// <param name="serviceImpl">An object implementing the server-side handling logic.</param>
    public static grpc::ServerServiceDefinition BindService(DynamicLocGroupApiBase serviceImpl)
    {
      return grpc::ServerServiceDefinition.CreateBuilder()
          .AddMethod(__Method_SendToGroup, serviceImpl.SendToGroup).Build();
    }

    /// <summary>Register service method implementations with a service binder. Useful when customizing the service binding logic.
    /// Note: this method is part of an experimental API that can change or be removed without any prior notice.</summary>
    /// <param name="serviceBinder">Service methods will be bound by calling <c>AddMethod</c> on this object.</param>
    /// <param name="serviceImpl">An object implementing the server-side handling logic.</param>
    public static void BindService(grpc::ServiceBinderBase serviceBinder, DynamicLocGroupApiBase serviceImpl)
    {
      serviceBinder.AddMethod(__Method_SendToGroup, serviceImpl.SendToGroup);
    }

  }
}
#endregion
