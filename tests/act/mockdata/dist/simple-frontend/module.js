/* [create-plugin] version: 7.6.0 */
/* [create-plugin] plugin: grafana-simplefrontend-panel@1.0.0 */
define(["@emotion/css","@grafana/data","@grafana/runtime","@grafana/ui","module","react"],(e,t,a,n,s,o)=>(()=>{"use strict";var r={89(t){t.exports=e},781(e){e.exports=t},531(e){e.exports=a},7(e){e.exports=n},308(e){e.exports=s},959(e){e.exports=o}};const i={};function l(e){const t=i[e];if(void 0!==t)return t.exports;const a=i[e]={exports:{}};return r[e](a,a.exports,l),a.exports}l.n=e=>{const t=e&&e.__esModule?()=>e.default:()=>e;return l.d(t,{a:t}),t},l.d=(e,t)=>{for(var a in t)l.o(t,a)&&!l.o(e,a)&&Object.defineProperty(e,a,{enumerable:!0,get:t[a]})},l.o=(e,t)=>Object.prototype.hasOwnProperty.call(e,t),l.r=e=>{Object.defineProperty(e,Symbol.toStringTag,{value:"Module"}),Object.defineProperty(e,"__esModule",{value:!0})};let p={};l.r(p),l.d(p,{plugin:()=>v});var u=l(308),d=l.n(u);l.p=d()&&d().uri?d().uri.slice(0,d().uri.lastIndexOf("/")+1):"public/plugins/grafana-simplefrontend-panel/";var c=l(781),m=l(959),f=l.n(m),g=l(89),x=l(7),h=l(531);const w=()=>({wrapper:g.css`
      font-family: Open Sans;
      position: relative;
    `,svg:g.css`
      position: absolute;
      top: 0;
      left: 0;
    `,textBox:g.css`
      position: absolute;
      bottom: 0;
      left: 0;
      padding: 10px;
    `}),v=new c.PanelPlugin(({options:e,data:t,width:a,height:n,fieldConfig:s,id:o})=>{const r=(0,x.useTheme2)(),i=(0,x.useStyles2)(w);return 0===t.series.length?f().createElement(h.PanelDataErrorView,{fieldConfig:s,panelId:o,data:t,needsStringField:!0}):f().createElement("div",{className:(0,g.cx)(i.wrapper,g.css`
          width: ${a}px;
          height: ${n}px;
        `)},f().createElement("svg",{className:i.svg,width:a,height:n,xmlns:"http://www.w3.org/2000/svg",xmlnsXlink:"http://www.w3.org/1999/xlink",viewBox:`-${a/2} -${n/2} ${a} ${n}`},f().createElement("g",null,f().createElement("circle",{"data-testid":"simple-panel-circle",style:{fill:r.colors.primary.main},r:100}))),f().createElement("div",{className:i.textBox},e.showSeriesCount&&f().createElement("div",{"data-testid":"simple-panel-series-counter"},"Number of series: ",t.series.length),f().createElement("div",null,"Text option value: ",e.text)))}).setPanelOptions(e=>e.addTextInput({path:"text",name:"Simple text option",description:"Description of panel option",defaultValue:"Default value of text input option"}).addBooleanSwitch({path:"showSeriesCount",name:"Show series counter",defaultValue:!1}).addRadio({path:"seriesCountSize",defaultValue:"sm",name:"Series counter size",settings:{options:[{value:"sm",label:"Small"},{value:"md",label:"Medium"},{value:"lg",label:"Large"}]},showIf:e=>e.showSeriesCount}));return p})());
//# sourceMappingURL=module.js.map