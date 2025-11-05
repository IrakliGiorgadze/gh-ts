
// Very small mock API backed by localStorage.
// Provides: users, tickets, comments, simple SLA indicator.

const DEMO_NS = 'it_demo_v1';

function _load(){
  const raw = localStorage.getItem(DEMO_NS);
  if(raw) return JSON.parse(raw);
  // seed
  const now = Date.now();
  const data = {
    users:[{id:'u1',email:'admin@example.com',name:'Admin',role:'admin',password:'admin'}],
    tickets:[
      {id:'T-1001', title:'VPN not connecting', description:'Fails on step 2', category:'Network', priority:'High',
       department:'Finance', status:'Open', assignee:'Admin', updatedAt: now-3600_000, createdAt: now-7200_000, comments:[{text:'Checking logs', createdAt: now-3500_000}]},
      {id:'T-1002', title:'Email quota exceeded', description:'Cannot send', category:'Software', priority:'Medium',
       department:'Sales', status:'Pending', assignee:'', updatedAt: now-1800_000, createdAt: now-5400_000, comments:[]},
      {id:'T-1003', title:'Laptop fan noise', description:'Very loud', category:'Hardware', priority:'Low',
       department:'HR', status:'New', assignee:'', updatedAt: now-600_000, createdAt: now-500_000, comments:[]}
    ],
    nextId:1004
  };
  localStorage.setItem(DEMO_NS, JSON.stringify(data));
  return data;
}

function _save(data){ localStorage.setItem(DEMO_NS, JSON.stringify(data)); }

const API = {
  async listTickets({q='', status='' }={}){
    const db = _load();
    let arr = db.tickets.slice().sort((a,b)=>b.updatedAt-a.updatedAt);
    if(q){ const needle = q.toLowerCase(); arr = arr.filter(t => (t.title+t.description).toLowerCase().includes(needle)); }
    if(status){ arr = arr.filter(t => t.status === status); }
    return arr;
  },
  async getStats(){
    const db = _load();
    const open = db.tickets.filter(t=>!['Resolved','Closed'].includes(t.status)).length;
    const resolved7d = db.tickets.filter(t=> t.status==='Resolved' && (Date.now()-t.updatedAt)<7*86400_000).length;
    const risk = db.tickets.filter(t=> ['High','Critical'].includes(t.priority) && !['Resolved','Closed'].includes(t.status)).length;
    return {open, resolved7d, risk};
  },
  async createTicket(t){
    const db = _load();
    const id = `T-${db.nextId++}`;
    const now = Date.now();
    const rec = {id, title:t.title, description:t.description||'', category:t.category||'Software',
      priority:t.priority||'Low', department:t.department||'', status:'New', assignee:'', updatedAt: now, createdAt: now, comments:[]};
    db.tickets.push(rec); _save(db); return rec;
  },
  async getTicket(id){
    const db = _load();
    return db.tickets.find(t=>t.id===id);
  },
  async updateTicket(id, patch){
    const db = _load();
    const t = db.tickets.find(t=>t.id===id);
    if(!t) throw new Error('Not found');
    Object.assign(t, patch); t.updatedAt = Date.now(); _save(db); return t;
  },
  async addComment(id, text){
    const db = _load();
    const t = db.tickets.find(t=>t.id===id);
    if(!t) throw new Error('Not found');
    t.comments = t.comments || [];
    t.comments.push({text, createdAt: Date.now()});
    t.updatedAt = Date.now();
    _save(db); return t;
  }
};

const Auth = {
  async login(email, password){
    const db = _load();
    const user = db.users.find(u=>u.email===email && u.password===password);
    if(user){ localStorage.setItem('demo_email', user.email); return true; }
    return False;
  },
  logout(){ localStorage.removeItem('demo_email'); }
};
