package main

import (
	"html/template"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/mobiledgex/edge-cloud/gensupport"
	"github.com/mobiledgex/edge-cloud/protogen"
)

type GenNotify struct {
	*generator.Generator
	support           gensupport.PluginSupport
	tmpl              *template.Template
	importSync        bool
	importLog         bool
	importContext     bool
	importOpentracing bool
}

func (g *GenNotify) Name() string {
	return "GenNotify"
}

func (g *GenNotify) Init(gen *generator.Generator) {
	g.Generator = gen
	g.tmpl = template.Must(template.New("notify").Parse(tmpl))
}

func (g *GenNotify) GenerateImports(file *generator.FileDescriptor) {
	g.support.PrintUsedImports(g.Generator)
	if g.importSync {
		g.PrintImport("", "sync")
	}
	g.PrintImport("", "github.com/gogo/protobuf/types")
	if g.importLog {
		g.PrintImport("", "github.com/mobiledgex/edge-cloud/log")
	}
	if g.importContext {
		g.PrintImport("", "context")
	}
	if g.importOpentracing {
		g.PrintImport("opentracing", "github.com/opentracing/opentracing-go")
	}
}

func (g *GenNotify) Generate(file *generator.FileDescriptor) {
	g.importSync = false
	g.importLog = false
	g.importContext = false
	g.importOpentracing = false

	g.support.InitFile()
	if !g.support.GenFile(*file.FileDescriptorProto.Name) {
		return
	}
	dogen := false
	for _, desc := range file.Messages() {
		if desc.File() != file.FileDescriptorProto {
			continue
		}
		if GetNotifyCache(desc.DescriptorProto) || GetNotifyMessage(desc.DescriptorProto) {
			dogen = true
			break
		}
	}
	if !dogen {
		return
	}

	g.P(gensupport.AutoGenComment)

	for _, desc := range file.Messages() {
		if desc.File() != file.FileDescriptorProto {
			continue
		}
		g.generateMessage(file, desc)
	}
	gensupport.RunParseCheck(g.Generator, file)
}

func (g *GenNotify) generateMessage(file *generator.FileDescriptor, desc *generator.Descriptor) {
	message := desc.DescriptorProto

	args := tmplArgs{
		Name: *message.Name,
	}
	if GetNotifyCache(message) {
		keyField := gensupport.GetMessageKey(message)
		if keyField == nil {
			g.Fail("message", *message.Name, "needs a unique key field named key of type ", *message.Name+"Key", "for option notify_cache")
		}
		args.KeyType = g.support.GoType(g.Generator, keyField)
		args.Cache = true
	} else if GetNotifyMessage(message) {
		args.Cache = false
	} else {
		return
	}
	args.NameType = g.support.FQTypeName(g.Generator, desc)
	args.CustomUpdate = GetNotifyCustomUpdate(message)
	args.RecvHook = GetNotifyRecvHook(message)
	args.Flush = GetNotifyFlush(message)
	args.FilterCloudletKey = GetNotifyFilterCloudletKey(message)
	args.PrintSendRecv = GetNotifyPrintSendRecv(message)
	g.tmpl.Execute(g.Generator.Buffer, args)

	g.importSync = true
	g.importContext = true
	g.importLog = true
	if args.PrintSendRecv {
		g.importOpentracing = true
	}
}

type tmplArgs struct {
	Name              string
	NameType          string
	KeyType           string
	Cache             bool
	Flush             bool
	CustomUpdate      bool
	RecvHook          bool
	FilterCloudletKey bool
	PrintSendRecv     bool
}

var tmpl = `
{{- if .Cache}}
type Send{{.Name}}Handler interface {
	GetAllKeys(ctx context.Context, keys map[{{.KeyType}}]context.Context)
	Get(key *{{.KeyType}}, buf *{{.NameType}}) bool
{{- if .FilterCloudletKey}}
	GetForCloudlet(key *edgeproto.CloudletKey, keys map[{{.KeyType}}]struct{})
{{- end}}
}

type Recv{{.Name}}Handler interface {
	Update(ctx context.Context, in *{{.NameType}}, rev int64)
	Delete(ctx context.Context, in *{{.NameType}}, rev int64)
	Prune(ctx context.Context, keys map[{{.KeyType}}]struct{})
	Flush(ctx context.Context, notifyId int64)
}
{{- else}}
type Recv{{.Name}}Handler interface {
	Recv(ctx context.Context, msg *{{.NameType}})
}
{{- end}}

type {{.Name}}Send struct {
	Name string
	MessageName string
{{- if .Cache}}
	handler Send{{.Name}}Handler
	Keys map[{{.KeyType}}]context.Context
	keysToSend map[{{.KeyType}}]context.Context
{{- else}}
	Data []*{{.NameType}}
	dataToSend []*{{.NameType}}
	Ctxs []context.Context
	ctxsToSend []context.Context
{{- end}}
	Mux sync.Mutex
	buf {{.NameType}}
	SendCount uint64
	sendrecv *SendRecv
}

{{- if .Cache}}
func New{{.Name}}Send(handler Send{{.Name}}Handler) *{{.Name}}Send {
{{- else}}
func New{{.Name}}Send() *{{.Name}}Send {
{{- end}}
	send := &{{.Name}}Send{}
	send.Name = "{{.Name}}"
	send.MessageName = proto.MessageName((*{{.NameType}})(nil))
{{- if .Cache}}
	send.handler = handler
	send.Keys = make(map[{{.KeyType}}]context.Context)
{{- else}}
	send.Data = make([]*{{.NameType}}, 0)
{{- end}}
	return send
}

func (s *{{.Name}}Send) SetSendRecv(sendrecv *SendRecv) {
	s.sendrecv = sendrecv
}

func (s *{{.Name}}Send) GetMessageName() string {
	return s.MessageName
}

func (s *{{.Name}}Send) GetName() string {
	return s.Name
}

func (s *{{.Name}}Send) GetSendCount() uint64 {
	return s.SendCount
}

{{- if .Cache}}
func (s *{{.Name}}Send) UpdateAll(ctx context.Context) {
	if !s.sendrecv.isRemoteWanted(s.MessageName) {
		return
	}
{{- if .CustomUpdate}}
	if !s.UpdateAllOk() { // to be implemented by hand
		return
	}
{{- end}}
	s.Mux.Lock()
	s.handler.GetAllKeys(ctx, s.Keys)
	s.Mux.Unlock()
}

func (s *{{.Name}}Send) Update(ctx context.Context, key *{{.KeyType}}, old *{{.NameType}}) {
	if !s.sendrecv.isRemoteWanted(s.MessageName) {
		return
	}
{{- if .CustomUpdate}}
	if !s.UpdateOk(ctx, key) { // to be implemented by hand
		return
	}
{{- end}}
	s.updateInternal(ctx, key)
}

func (s *{{.Name}}Send) updateInternal(ctx context.Context, key *{{.KeyType}}) {
	s.Mux.Lock()
{{- if .PrintSendRecv}}
	log.SpanLog(ctx, log.DebugLevelNotify, "updateInternal {{.Name}}", "key", key)
{{- end}}
	s.Keys[*key] = ctx
	s.Mux.Unlock()
	s.sendrecv.wakeup()
}
{{- else}}
func (s *{{.Name}}Send) UpdateAll(ctx context.Context) {}

func (s *{{.Name}}Send) Update(ctx context.Context, msg *{{.NameType}}) bool {
{{- if .CustomUpdate}}
	if !s.UpdateOk(ctx, msg) { // to be implemented by hand
		return false
	}
{{- end}}
	s.Mux.Lock()
	s.Data = append(s.Data, msg)
	s.Ctxs = append(s.Ctxs, ctx)
	s.Mux.Unlock()
	s.sendrecv.wakeup()
	return true
}
{{- end}}

func (s *{{.Name}}Send) Send(stream StreamNotify, notice *edgeproto.Notice, peer string) error {
	s.Mux.Lock()
{{- if .Cache}}
	keys := s.keysToSend
	s.keysToSend = nil
{{- else}}
	data := s.dataToSend
	s.dataToSend = nil
	ctxs := s.ctxsToSend
	s.ctxsToSend = nil
{{- end}}
	s.Mux.Unlock()

{{- if .Cache}}
	for key, ctx := range keys {
		found := s.handler.Get(&key, &s.buf)
		if found {
			notice.Action = edgeproto.NoticeAction_UPDATE
		} else {
			notice.Action = edgeproto.NoticeAction_DELETE
			s.buf.Key = key
		}
		any, err := types.MarshalAny(&s.buf)
{{- else}}
	for ii, msg := range data {
		any, err := types.MarshalAny(msg)
		ctx := ctxs[ii]
{{- end}}
		if err != nil {
			s.sendrecv.stats.MarshalErrors++
			err = nil
			continue
		}
		notice.Any = *any
		notice.Span = log.SpanToString(ctx)
{{- if .PrintSendRecv}}
		log.SpanLog(ctx, log.DebugLevelNotify,
			fmt.Sprintf("%s send {{.Name}}", s.sendrecv.cliserv),
			"peer", peer,
{{- if .Cache}}
			"action", notice.Action,
			"key", key)
{{- else}}
			"msg", msg)
{{- end}}
{{- end}}
		err = stream.Send(notice)
		if err != nil {
			s.sendrecv.stats.SendErrors++
			return err
		}
		s.sendrecv.stats.Send++
		// object specific counter
		s.SendCount++
	}
	return nil
}

func (s *{{.Name}}Send) PrepData() bool {
	s.Mux.Lock()
	defer s.Mux.Unlock()
{{- if .Cache}}
	if len(s.Keys) > 0 {
		s.keysToSend = s.Keys
		s.Keys = make(map[{{.KeyType}}]context.Context)
		return true
	}
{{- else}}
	if len(s.Data) > 0 {
		s.dataToSend = s.Data
		s.Data = make([]*{{.NameType}}, 0)
		s.ctxsToSend = s.Ctxs
		s.Ctxs = make([]context.Context, 0)
		return true
	}
{{- end}}
	return false
}

// Server accepts multiple clients so needs to track multiple
// peers to send to.
type {{.Name}}SendMany struct {
{{- if .Cache}}
	handler Send{{.Name}}Handler
{{- end}}
	Mux sync.Mutex
	sends map[string]*{{.Name}}Send
}

{{- if .Cache}}
func New{{.Name}}SendMany(handler Send{{.Name}}Handler) *{{.Name}}SendMany {
{{- else}}
func New{{.Name}}SendMany() *{{.Name}}SendMany {
{{- end}}
	s := &{{.Name}}SendMany{}
{{- if .Cache}}
	s.handler = handler
{{- end}}
	s.sends = make(map[string]*{{.Name}}Send)
	return s
}

func (s *{{.Name}}SendMany) NewSend(peerAddr string) NotifySend {
{{- if .Cache}}
	send := New{{.Name}}Send(s.handler)
{{- else}}
	send := New{{.Name}}Send()
{{- end}}
	s.Mux.Lock()
	s.sends[peerAddr] = send
	s.Mux.Unlock()
	return send
}

func (s *{{.Name}}SendMany) DoneSend(peerAddr string, send NotifySend) {
	asend, ok := send.(*{{.Name}}Send)
	if !ok {
		return
	}
	// another connection may come from the same client so remove
	// only if it matches
	s.Mux.Lock()
	if remove, _ := s.sends[peerAddr]; remove == asend {
		delete(s.sends, peerAddr)
	}
	s.Mux.Unlock()
}

{{- if .Cache}}
func (s *{{.Name}}SendMany) Update(ctx context.Context, key *{{.KeyType}}, old *{{.NameType}}) {
	s.Mux.Lock()
	defer s.Mux.Unlock()
	for _, send := range s.sends {
		send.Update(ctx, key, old)
	}
}
{{- else}}
func (s *{{.Name}}SendMany) Update(ctx context.Context, msg *{{.NameType}}) int {
	s.Mux.Lock()
	defer s.Mux.Unlock()
	count := 0
	for _, send := range s.sends {
		if send.Update(ctx, msg) {
			count++
		}
	}
	return count
}
{{- end}}

type {{.Name}}Recv struct {
	Name string
	MessageName string
	handler Recv{{.Name}}Handler
{{- if .Cache}}
	sendAllKeys map[{{.KeyType}}]struct{}
{{- end}}
	Mux sync.Mutex
	buf {{.NameType}}
	RecvCount uint64
	sendrecv *SendRecv
}

func New{{.Name}}Recv(handler Recv{{.Name}}Handler) *{{.Name}}Recv {
	recv := &{{.Name}}Recv{}
	recv.Name = "{{.Name}}"
	recv.MessageName = proto.MessageName((*{{.NameType}})(nil))
	recv.handler = handler
	return recv
}

func (s *{{.Name}}Recv) SetSendRecv(sendrecv *SendRecv) {
	s.sendrecv = sendrecv
}

func (s *{{.Name}}Recv) GetMessageName() string {
	return s.MessageName
}

func (s *{{.Name}}Recv) GetName() string {
	return s.Name
}

func (s *{{.Name}}Recv) GetRecvCount() uint64 {
	return s.RecvCount
}

func (s *{{.Name}}Recv) Recv(ctx context.Context, notice *edgeproto.Notice, notifyId int64, peer string) {
{{- if .PrintSendRecv}}
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		span.SetTag("objtype", "{{.Name}}")
	}
{{- end}}

	buf := &{{.NameType}}{}
	err := types.UnmarshalAny(&notice.Any, buf)
	if err != nil {
		s.sendrecv.stats.UnmarshalErrors++
{{- if .PrintSendRecv}}
		log.SpanLog(ctx, log.DebugLevelNotify, "err", err)
{{- end}}
		return
	}
{{- if .Flush}}
	buf.NotifyId = notifyId
{{- end}}
{{- if .PrintSendRecv}}
	if span != nil {
{{- if .Cache}}
		span.SetTag("key", buf.Key)
{{- else}}
		span.SetTag("msg", buf)
{{- end}}
	}
{{- end}}
{{- if .Cache}}
	if notice.Action == edgeproto.NoticeAction_UPDATE {
		s.handler.Update(ctx, buf, 0)
		s.Mux.Lock()
		if s.sendAllKeys != nil {
			s.sendAllKeys[buf.Key] = struct{}{}
		}
		s.Mux.Unlock()
	} else if notice.Action == edgeproto.NoticeAction_DELETE {
		s.handler.Delete(ctx, buf, 0)
	}
{{- else}}
	s.handler.Recv(ctx, buf)
{{- end}}
	s.sendrecv.stats.Recv++
	// object specific counter
	s.RecvCount++
{{- if .RecvHook}}
	s.RecvHook(ctx, notice, buf, peer) // to be implemented by hand
{{- end}}
}

func (s *{{.Name}}Recv) RecvAllStart() {
{{- if .Cache}}
	s.Mux.Lock()
	defer s.Mux.Unlock()
	s.sendAllKeys = make(map[{{.KeyType}}]struct{})
{{- end}}
}

func (s *{{.Name}}Recv) RecvAllEnd(ctx context.Context, cleanup Cleanup) {
{{- if .Cache}}
	s.Mux.Lock()
	validKeys := s.sendAllKeys
	s.sendAllKeys = nil
	s.Mux.Unlock()
	if cleanup == CleanupPrune {
		s.handler.Prune(ctx, validKeys)
	}
{{- end}}
}

{{- if .Cache}}
func (s *{{.Name}}Recv) Flush(ctx context.Context, notifyId int64) {
	s.handler.Flush(ctx, notifyId)
}
{{- end}}

type {{.Name}}RecvMany struct {
	handler Recv{{.Name}}Handler
}

func New{{.Name}}RecvMany(handler Recv{{.Name}}Handler) *{{.Name}}RecvMany {
	s := &{{.Name}}RecvMany{}
	s.handler = handler
	return s
}

func (s *{{.Name}}RecvMany) NewRecv() NotifyRecv {
	recv := New{{.Name}}Recv(s.handler)
	return recv
}

func (s *{{.Name}}RecvMany) Flush(ctx context.Context, notifyId int64) {
{{- if .Cache}}
	s.handler.Flush(ctx, notifyId)
{{- end}}
}

{{- if .Cache}}
func (mgr *ServerMgr) RegisterSend{{.Name}}Cache(cache *edgeproto.{{.Name}}Cache) {
	send := New{{.Name}}SendMany(cache)
	mgr.RegisterSend(send)
	cache.SetNotifyCb(send.Update)
}

func (mgr *ServerMgr) RegisterRecv{{.Name}}Cache(cache *edgeproto.{{.Name}}Cache) {
	recv := New{{.Name}}RecvMany(cache)
	mgr.RegisterRecv(recv)
}

func (s *Client) RegisterSend{{.Name}}Cache(cache *edgeproto.{{.Name}}Cache) {
	send := New{{.Name}}Send(cache)
	s.RegisterSend(send)
	cache.SetNotifyCb(send.Update)
}

func (s *Client) RegisterRecv{{.Name}}Cache(cache *edgeproto.{{.Name}}Cache) {
	recv := New{{.Name}}Recv(cache)
	s.RegisterRecv(recv)
}
{{- end}}

`

func GetNotifyCache(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_NotifyCache, false)
}

func GetNotifyFlush(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_NotifyFlush, false)
}

func GetNotifyMessage(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_NotifyMessage, false)
}

func GetNotifyCustomUpdate(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_NotifyCustomUpdate, false)
}

func GetNotifyRecvHook(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_NotifyRecvHook, false)
}

func GetNotifyFilterCloudletKey(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_NotifyFilterCloudletKey, false)
}

func GetNotifyPrintSendRecv(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_NotifyPrintSendRecv, true)
}
