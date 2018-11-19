// <auto-generated>
//     Generated by the protocol buffer compiler.  DO NOT EDIT!
//     source: appcommon.proto
// </auto-generated>
#pragma warning disable 1591, 0612, 3021
#region Designer generated code

using pb = global::Google.Protobuf;
using pbc = global::Google.Protobuf.Collections;
using pbr = global::Google.Protobuf.Reflection;
using scg = global::System.Collections.Generic;
namespace DistributedMatchEngine {

  /// <summary>Holder for reflection information generated from appcommon.proto</summary>
  public static partial class AppcommonReflection {

    #region Descriptor
    /// <summary>File descriptor for appcommon.proto</summary>
    public static pbr::FileDescriptor Descriptor {
      get { return descriptor; }
    }
    private static pbr::FileDescriptor descriptor;

    static AppcommonReflection() {
      byte[] descriptorData = global::System.Convert.FromBase64String(
          string.Concat(
            "Cg9hcHBjb21tb24ucHJvdG8SGGRpc3RyaWJ1dGVkX21hdGNoX2VuZ2luZSKQ",
            "AQoHQXBwUG9ydBIvCgVwcm90bxgBIAEoDjIgLmRpc3RyaWJ1dGVkX21hdGNo",
            "X2VuZ2luZS5MUHJvdG8SFQoNaW50ZXJuYWxfcG9ydBgCIAEoBRITCgtwdWJs",
            "aWNfcG9ydBgDIAEoBRITCgtwdWJsaWNfcGF0aBgEIAEoCRITCgtGUUROX3By",
            "ZWZpeBgFIAEoCSpJCgZMUHJvdG8SEQoNTFByb3RvVW5rbm93bhAAEg0KCUxQ",
            "cm90b1RDUBABEg0KCUxQcm90b1VEUBACEg4KCkxQcm90b0hUVFAQA2IGcHJv",
            "dG8z"));
      descriptor = pbr::FileDescriptor.FromGeneratedCode(descriptorData,
          new pbr::FileDescriptor[] { },
          new pbr::GeneratedClrTypeInfo(new[] {typeof(global::DistributedMatchEngine.LProto), }, new pbr::GeneratedClrTypeInfo[] {
            new pbr::GeneratedClrTypeInfo(typeof(global::DistributedMatchEngine.AppPort), global::DistributedMatchEngine.AppPort.Parser, new[]{ "Proto", "InternalPort", "PublicPort", "PublicPath", "FQDNPrefix" }, null, null, null)
          }));
    }
    #endregion

  }
  #region Enums
  /// <summary>
  /// LProto indicates which protocol to use for accessing an application on a particular port. This is required by Kubernetes for port mapping.
  /// </summary>
  public enum LProto {
    /// <summary>
    /// Unknown protocol
    /// </summary>
    [pbr::OriginalName("LProtoUnknown")] Unknown = 0,
    /// <summary>
    /// TCP (L4) protocol
    /// </summary>
    [pbr::OriginalName("LProtoTCP")] Tcp = 1,
    /// <summary>
    /// UDP (L4) protocol
    /// </summary>
    [pbr::OriginalName("LProtoUDP")] Udp = 2,
    /// <summary>
    /// HTTP (L7 tcp) protocol
    /// </summary>
    [pbr::OriginalName("LProtoHTTP")] Http = 3,
  }

  #endregion

  #region Messages
  /// <summary>
  /// AppPort describes an L4 or L7 public access port/path mapping. This is used to track external to internal mappings for access via a shared load balancer or reverse proxy.
  /// </summary>
  public sealed partial class AppPort : pb::IMessage<AppPort> {
    private static readonly pb::MessageParser<AppPort> _parser = new pb::MessageParser<AppPort>(() => new AppPort());
    private pb::UnknownFieldSet _unknownFields;
    [global::System.Diagnostics.DebuggerNonUserCodeAttribute]
    public static pb::MessageParser<AppPort> Parser { get { return _parser; } }

    [global::System.Diagnostics.DebuggerNonUserCodeAttribute]
    public static pbr::MessageDescriptor Descriptor {
      get { return global::DistributedMatchEngine.AppcommonReflection.Descriptor.MessageTypes[0]; }
    }

    [global::System.Diagnostics.DebuggerNonUserCodeAttribute]
    pbr::MessageDescriptor pb::IMessage.Descriptor {
      get { return Descriptor; }
    }

    [global::System.Diagnostics.DebuggerNonUserCodeAttribute]
    public AppPort() {
      OnConstruction();
    }

    partial void OnConstruction();

    [global::System.Diagnostics.DebuggerNonUserCodeAttribute]
    public AppPort(AppPort other) : this() {
      proto_ = other.proto_;
      internalPort_ = other.internalPort_;
      publicPort_ = other.publicPort_;
      publicPath_ = other.publicPath_;
      fQDNPrefix_ = other.fQDNPrefix_;
      _unknownFields = pb::UnknownFieldSet.Clone(other._unknownFields);
    }

    [global::System.Diagnostics.DebuggerNonUserCodeAttribute]
    public AppPort Clone() {
      return new AppPort(this);
    }

    /// <summary>Field number for the "proto" field.</summary>
    public const int ProtoFieldNumber = 1;
    private global::DistributedMatchEngine.LProto proto_ = 0;
    /// <summary>
    /// TCP (L4), UDP (L4), or HTTP (L7) protocol
    /// </summary>
    [global::System.Diagnostics.DebuggerNonUserCodeAttribute]
    public global::DistributedMatchEngine.LProto Proto {
      get { return proto_; }
      set {
        proto_ = value;
      }
    }

    /// <summary>Field number for the "internal_port" field.</summary>
    public const int InternalPortFieldNumber = 2;
    private int internalPort_;
    /// <summary>
    /// Container port
    /// </summary>
    [global::System.Diagnostics.DebuggerNonUserCodeAttribute]
    public int InternalPort {
      get { return internalPort_; }
      set {
        internalPort_ = value;
      }
    }

    /// <summary>Field number for the "public_port" field.</summary>
    public const int PublicPortFieldNumber = 3;
    private int publicPort_;
    /// <summary>
    /// Public facing port for TCP/UDP (may be mapped on shared LB reverse proxy)
    /// </summary>
    [global::System.Diagnostics.DebuggerNonUserCodeAttribute]
    public int PublicPort {
      get { return publicPort_; }
      set {
        publicPort_ = value;
      }
    }

    /// <summary>Field number for the "public_path" field.</summary>
    public const int PublicPathFieldNumber = 4;
    private string publicPath_ = "";
    /// <summary>
    /// Public facing path for HTTP L7 access.
    /// </summary>
    [global::System.Diagnostics.DebuggerNonUserCodeAttribute]
    public string PublicPath {
      get { return publicPath_; }
      set {
        publicPath_ = pb::ProtoPreconditions.CheckNotNull(value, "value");
      }
    }

    /// <summary>Field number for the "FQDN_prefix" field.</summary>
    public const int FQDNPrefixFieldNumber = 5;
    private string fQDNPrefix_ = "";
    /// <summary>
    /// FQDN prefix to append to base FQDN in FindCloudlet response. May be empty.
    /// </summary>
    [global::System.Diagnostics.DebuggerNonUserCodeAttribute]
    public string FQDNPrefix {
      get { return fQDNPrefix_; }
      set {
        fQDNPrefix_ = pb::ProtoPreconditions.CheckNotNull(value, "value");
      }
    }

    [global::System.Diagnostics.DebuggerNonUserCodeAttribute]
    public override bool Equals(object other) {
      return Equals(other as AppPort);
    }

    [global::System.Diagnostics.DebuggerNonUserCodeAttribute]
    public bool Equals(AppPort other) {
      if (ReferenceEquals(other, null)) {
        return false;
      }
      if (ReferenceEquals(other, this)) {
        return true;
      }
      if (Proto != other.Proto) return false;
      if (InternalPort != other.InternalPort) return false;
      if (PublicPort != other.PublicPort) return false;
      if (PublicPath != other.PublicPath) return false;
      if (FQDNPrefix != other.FQDNPrefix) return false;
      return Equals(_unknownFields, other._unknownFields);
    }

    [global::System.Diagnostics.DebuggerNonUserCodeAttribute]
    public override int GetHashCode() {
      int hash = 1;
      if (Proto != 0) hash ^= Proto.GetHashCode();
      if (InternalPort != 0) hash ^= InternalPort.GetHashCode();
      if (PublicPort != 0) hash ^= PublicPort.GetHashCode();
      if (PublicPath.Length != 0) hash ^= PublicPath.GetHashCode();
      if (FQDNPrefix.Length != 0) hash ^= FQDNPrefix.GetHashCode();
      if (_unknownFields != null) {
        hash ^= _unknownFields.GetHashCode();
      }
      return hash;
    }

    [global::System.Diagnostics.DebuggerNonUserCodeAttribute]
    public override string ToString() {
      return pb::JsonFormatter.ToDiagnosticString(this);
    }

    [global::System.Diagnostics.DebuggerNonUserCodeAttribute]
    public void WriteTo(pb::CodedOutputStream output) {
      if (Proto != 0) {
        output.WriteRawTag(8);
        output.WriteEnum((int) Proto);
      }
      if (InternalPort != 0) {
        output.WriteRawTag(16);
        output.WriteInt32(InternalPort);
      }
      if (PublicPort != 0) {
        output.WriteRawTag(24);
        output.WriteInt32(PublicPort);
      }
      if (PublicPath.Length != 0) {
        output.WriteRawTag(34);
        output.WriteString(PublicPath);
      }
      if (FQDNPrefix.Length != 0) {
        output.WriteRawTag(42);
        output.WriteString(FQDNPrefix);
      }
      if (_unknownFields != null) {
        _unknownFields.WriteTo(output);
      }
    }

    [global::System.Diagnostics.DebuggerNonUserCodeAttribute]
    public int CalculateSize() {
      int size = 0;
      if (Proto != 0) {
        size += 1 + pb::CodedOutputStream.ComputeEnumSize((int) Proto);
      }
      if (InternalPort != 0) {
        size += 1 + pb::CodedOutputStream.ComputeInt32Size(InternalPort);
      }
      if (PublicPort != 0) {
        size += 1 + pb::CodedOutputStream.ComputeInt32Size(PublicPort);
      }
      if (PublicPath.Length != 0) {
        size += 1 + pb::CodedOutputStream.ComputeStringSize(PublicPath);
      }
      if (FQDNPrefix.Length != 0) {
        size += 1 + pb::CodedOutputStream.ComputeStringSize(FQDNPrefix);
      }
      if (_unknownFields != null) {
        size += _unknownFields.CalculateSize();
      }
      return size;
    }

    [global::System.Diagnostics.DebuggerNonUserCodeAttribute]
    public void MergeFrom(AppPort other) {
      if (other == null) {
        return;
      }
      if (other.Proto != 0) {
        Proto = other.Proto;
      }
      if (other.InternalPort != 0) {
        InternalPort = other.InternalPort;
      }
      if (other.PublicPort != 0) {
        PublicPort = other.PublicPort;
      }
      if (other.PublicPath.Length != 0) {
        PublicPath = other.PublicPath;
      }
      if (other.FQDNPrefix.Length != 0) {
        FQDNPrefix = other.FQDNPrefix;
      }
      _unknownFields = pb::UnknownFieldSet.MergeFrom(_unknownFields, other._unknownFields);
    }

    [global::System.Diagnostics.DebuggerNonUserCodeAttribute]
    public void MergeFrom(pb::CodedInputStream input) {
      uint tag;
      while ((tag = input.ReadTag()) != 0) {
        switch(tag) {
          default:
            _unknownFields = pb::UnknownFieldSet.MergeFieldFrom(_unknownFields, input);
            break;
          case 8: {
            proto_ = (global::DistributedMatchEngine.LProto) input.ReadEnum();
            break;
          }
          case 16: {
            InternalPort = input.ReadInt32();
            break;
          }
          case 24: {
            PublicPort = input.ReadInt32();
            break;
          }
          case 34: {
            PublicPath = input.ReadString();
            break;
          }
          case 42: {
            FQDNPrefix = input.ReadString();
            break;
          }
        }
      }
    }

  }

  #endregion

}

#endregion Designer generated code
