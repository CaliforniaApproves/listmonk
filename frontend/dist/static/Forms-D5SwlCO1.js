import{n as a,V as l,m as n}from"./index-DGtM-77e.js";import{C as o}from"./CodeEditor-H_ctveuk.js";const p=l.extend({name:"ListForm",components:{"code-editor":o},data(){return{checked:[],html:"",selectedRedirectURL:""}},methods:{escapeAttr(s){return String(s).replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;").replace(/"/g,"&quot;").replace(/'/g,"&#39;")},renderHTML(){let s=`<form method="post" action="${this.serverConfig.root_url}/subscription/form" class="listmonk-form">
  <div>
    <h3>${this.$t("public.sub")}</h3>
    <input type="hidden" name="nonce" />
`;this.selectedRedirectURL&&(s+=`    <input type="hidden" name="next" value="${this.escapeAttr(this.selectedRedirectURL)}" />
`),s+=`
    <p><input type="email" name="email" required placeholder="${this.$t("subscribers.email")}" /></p>
    <p><input type="text" name="name" placeholder="${this.$t("public.subName")}" /></p>

`,this.checked.forEach(e=>{const t=this.publicLists[parseInt(e,10)];s+=`    <p>
      <input id="${t.uuid.substr(0,5)}" type="checkbox" name="l" checked value="${t.uuid}" />
      <label for="${t.uuid.substr(0,5)}">${t.name}</label>
`,t.description&&(s+=`      <br />
      <span>${t.description}</span>
`),s+=`    </p>
`}),this.serverConfig.public_subscription.captcha_enabled&&(this.serverConfig.public_subscription.captcha_provider==="altcha"?s+=`
    <altcha-widget challengeurl="${this.serverConfig.root_url}/api/public/captcha/altcha"></altcha-widget>
    <script type="module" src="${this.serverConfig.root_url}/public/static/altcha.umd.js" async defer><\/script>
`:this.serverConfig.public_subscription.captcha_provider==="hcaptcha"&&(s+=`
    <div class="h-captcha" data-sitekey="${this.serverConfig.public_subscription.captcha_key}"></div>
    <script src="https://js.hcaptcha.com/1/api.js" async defer><\/script>
`)),s+=`
    <input type="submit" value="${this.$t("public.sub")} " />
  </div>
</form>`,this.html=s}},computed:{...n(["loading","lists","serverConfig"]),publicLists(){return this.lists.results?this.lists.results.filter(s=>s.type==="public"):[]},redirectURLs(){const s=this.serverConfig.public_subscription?this.serverConfig.public_subscription.redirect_urls:[];return Array.isArray(s)?s:[]}},watch:{checked(){this.renderHTML()},selectedRedirectURL(){this.renderHTML()}}});var d=function(){var e=this,t=e._self._c;return e._self._setupProxy,t("section",{staticClass:"forms content relative"},[t("h1",{staticClass:"title is-4"},[e._v(" "+e._s(e.$t("forms.title"))+" ")]),t("hr"),e.loading.lists?t("b-loading",{attrs:{active:e.loading.lists,"is-full-page":!1}}):e.publicLists.length===0?t("p",[e._v(" "+e._s(e.$t("forms.noPublicLists"))+" ")]):e.publicLists.length>0?t("div",{staticClass:"columns"},[t("div",{staticClass:"column is-4"},[t("h4",[e._v(e._s(e.$t("forms.publicLists")))]),t("p",[e._v(e._s(e.$t("forms.selectHelp")))]),t("b-loading",{attrs:{active:e.loading.lists,"is-full-page":!1}}),t("ul",{staticClass:"no",attrs:{"data-cy":"lists"}},e._l(e.publicLists,function(i,r){return t("li",{key:i.id},[t("b-checkbox",{attrs:{"native-value":r},model:{value:e.checked,callback:function(c){e.checked=c},expression:"checked"}},[e._v(" "+e._s(i.name)+" ")])],1)}),0),e.serverConfig.public_subscription.enabled?[t("hr"),t("h4",[e._v(e._s(e.$t("forms.publicSubPage")))]),t("p",[t("a",{attrs:{href:`${e.serverConfig.root_url}/subscription/form`,target:"_blank",rel:"noopener noreferer","data-cy":"url"}},[e._v(" "+e._s(e.serverConfig.root_url)+"/subscription/form ")])])]:e._e(),t("hr"),t("h4",[e._v(e._s(e.$t("forms.redirectURL")))]),t("p",{staticClass:"is-size-7 has-text-grey"},[e._v(" "+e._s(e.$t("forms.redirectURLHelp"))+" ")]),e.redirectURLs.length>0?t("ul",{staticClass:"no",attrs:{"data-cy":"redirect-urls"}},[t("li",[t("b-radio",{attrs:{"native-value":""},model:{value:e.selectedRedirectURL,callback:function(i){e.selectedRedirectURL=i},expression:"selectedRedirectURL"}},[e._v(" "+e._s(e.$t("globals.terms.none"))+" ")])],1),e._l(e.redirectURLs,function(i){return t("li",{key:i},[t("b-radio",{attrs:{"native-value":i},model:{value:e.selectedRedirectURL,callback:function(r){e.selectedRedirectURL=r},expression:"selectedRedirectURL"}},[e._v(" "+e._s(i)+" ")])],1)})],2):e._e()],2),t("div",{staticClass:"column",attrs:{"data-cy":"form"}},[t("h4",[e._v(e._s(e.$t("forms.formHTML")))]),t("p",[e._v(" "+e._s(e.$t("forms.formHTMLHelp"))+" ")]),e.checked.length>0?t("code-editor",{attrs:{lang:"html",disabled:""},model:{value:e.html,callback:function(i){e.html=i},expression:"html"}}):e._e()],1)]):e._e()],1)},u=[],h=a(p,d,u);const f=h.exports;export{f as default};
