/* [create-plugin] version: 6.1.0 */
/* [create-plugin] plugin: grafana-simplefrontendyarn-panel@1.0.0 */
define(["@grafana/ui","@emotion/css","module","@grafana/runtime","@grafana/data","react"],(e,t,a,n,r,i)=>(()=>{"use strict";var l={7:t=>{t.exports=e},89:e=>{e.exports=t},308:e=>{e.exports=a},531:e=>{e.exports=n},781:e=>{e.exports=r},959:e=>{e.exports=i}},o={};function s(e){var t=o[e];if(void 0!==t)return t.exports;var a=o[e]={exports:{}};return l[e](a,a.exports,s),a.exports}s.n=e=>{var t=e&&e.__esModule?()=>e.default:()=>e;return s.d(t,{a:t}),t},s.d=(e,t)=>{for(var a in t)s.o(t,a)&&!s.o(e,a)&&Object.defineProperty(e,a,{enumerable:!0,get:t[a]})},s.o=(e,t)=>Object.prototype.hasOwnProperty.call(e,t),s.r=e=>{"undefined"!=typeof Symbol&&Symbol.toStringTag&&Object.defineProperty(e,Symbol.toStringTag,{value:"Module"}),Object.defineProperty(e,"__esModule",{value:!0})},s.p="public/plugins/grafana-simplefrontendyarn-panel/";var p={};s.r(p),s.d(p,{plugin:()=>w});var u=s(308),d=s.n(u);s.p=d()&&d().uri?d().uri.slice(0,d().uri.lastIndexOf("/")+1):"public/plugins/grafana-simplefrontendyarn-panel/";var c=s(781),m=s(959),f=s.n(m),g=s(89),x=s(7),v=s(531);const h=()=>({wrapper:g.css`
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
    `}),w=new c.PanelPlugin(({options:e,data:t,width:a,height:n,fieldConfig:r,id:i})=>{const l=(0,x.useTheme2)(),o=(0,x.useStyles2)(h);return 0===t.series.length?f().createElement(v.PanelDataErrorView,{fieldConfig:r,panelId:i,data:t,needsStringField:!0}):f().createElement("div",{className:(0,g.cx)(o.wrapper,g.css`
          width: ${a}px;
          height: ${n}px;
        `)},f().createElement("svg",{className:o.svg,width:a,height:n,xmlns:"http://www.w3.org/2000/svg",xmlnsXlink:"http://www.w3.org/1999/xlink",viewBox:`-${a/2} -${n/2} ${a} ${n}`},f().createElement("g",null,f().createElement("circle",{"data-testid":"simple-panel-circle",style:{fill:l.colors.primary.main},r:100}))),f().createElement("div",{className:o.textBox},e.showSeriesCount&&f().createElement("div",{"data-testid":"simple-panel-series-counter"},"Number of series: ",t.series.length),f().createElement("div",null,"Text option value: ",e.text)))}).setPanelOptions(e=>e.addTextInput({path:"text",name:"Simple text option",description:"Description of panel option",defaultValue:"Default value of text input option"}).addBooleanSwitch({path:"showSeriesCount",name:"Show series counter",defaultValue:!1}).addRadio({path:"seriesCountSize",defaultValue:"sm",name:"Series counter size",settings:{options:[{value:"sm",label:"Small"},{value:"md",label:"Medium"},{value:"lg",label:"Large"}]},showIf:e=>e.showSeriesCount}));return p})());
//# sourceMappingURL=module.js.map