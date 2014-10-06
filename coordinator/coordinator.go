package coordinator

import (
    "encoding/json"
	log "code.google.com/p/log4go"
	"github.com/megamsys/megamd/app"
	"github.com/megamsys/megamd/provisioner/chef"
	"github.com/megamsys/megamd/provisioner/docker"
	"github.com/megamsys/megamd/iaas/ec2"
	"github.com/megamsys/megamd/iaas/google"
	"github.com/megamsys/megamd/iaas/hp"
	"github.com/megamsys/megamd/iaas/gogrid"
	"github.com/megamsys/megamd/iaas/opennebula"
	"github.com/megamsys/megamd/iaas/profitbricks"
)

type Coordinator struct {
	//RequestHandler(f func(*Message), name ...string) (Handler, error)
	//EventsHandler(f func(*Message), name ...string) (Handler, error)
}

type Message struct {
	Id    string  `json:"id"`
}

func init() {
	ec2.Init()
	google.Init()
	hp.Init()
	gogrid.Init()
	opennebula.Init()
	profitbricks.Init()
	chef.Init()
	docker.Init()
}

func NewCoordinator(chann []byte, queue string) {
	log.Info("Handling coordinator message %v", string(chann))
	
	switch queue {
	case "cloudstandup":
	      requestHandler(chann)
	      break;
	case "Events":
	      eventsHandler(chann)
	      break;      
}
}
	
func requestHandler(chann []byte) {
	    m := &Message{} 
	    parse_err := json.Unmarshal(chann, &m)
	    if parse_err != nil {
	    	log.Error("Error: Message parsing error:\n%s.", parse_err)
			return
	    }
        request := app.Request{Id: m.Id}
        req, err := request.Get(m.Id)
		if err != nil {
			log.Error("Error: Riak didn't cooperate:\n%s.", err)
			return
		}
	   switch req.ReqType {
	   case "create":
	       	   assemblies := app.Assemblies{Id: req.AssembliesId }
               asm, err := assemblies.Get(req.AssembliesId)
		       if err != nil {
			         log.Error("Error: Riak didn't cooperate:\n%s.", err)
			         return
		         }
		       for i := range asm.Assemblies {
		       	if len(asm.Assemblies[i]) > 1 {
		             assemblyID := asm.Assemblies[i]
		             assembly := app.Assembly{Id: assemblyID }
                     res, err := assembly.Get(assemblyID)
		             if err != nil {
			            log.Error("Error: Riak didn't cooperate:\n%s.", err)
			            return
		              }
		          
		             go app.LaunchApp(res)
	             }
		       	}	
		}
}	
	   
func eventsHandler(chann []byte) {
	
	
}	   
	