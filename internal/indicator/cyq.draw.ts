
function drawCYQItem(thiskline: kline, containner: HTMLDivElement, index: number|undefined){

  if(index == undefined) index = thiskline.data.full_klines.length - 1

  if(containner.querySelector('canvas')){
    containner.removeChild(containner.querySelector('canvas')!)
  }

  let canvas = document.createElement('canvas')
  let width = thiskline.options.cyq_option.width - 10
  canvas.width = width
  canvas.height = thiskline.common_data.main_height - thiskline.common_data.top_message_height
  canvas.style.marginLeft = '10px'
  canvas.style.marginTop = thiskline.common_data.top_message_height + 'px'
  containner.appendChild(canvas)

  let ctx = canvas.getContext('2d') as CanvasRenderingContext2D

  let cyq_info_html:HTMLDivElement|null = null;
  if(containner.querySelector('.quotechart2022_c_cyq_info')){
    containner.removeChild(containner.querySelector('.quotechart2022_c_cyq_info')!)
  }
  cyq_info_html = document.createElement('div')
  cyq_info_html.className = 'quotechart2022_c_cyq_info'
  cyq_info_html.style.height = (thiskline.options.height - thiskline.common_data.main_height) + 'px'
  containner.appendChild(cyq_info_html)

  //@ts-ignore
  let cm1 = new CYQCalculator(thiskline.data.full_klines)
  let cm_result = cm1.calc(index)
  
  let y_max = thiskline.common_data.y_max
  // console.info(y_max)
  
  let y_scale = thiskline.common_data.y_scale

  let x_max:number = max(cm_result.x)!
  let x_scale = (thiskline.options.cyq_option.width - 50) / x_max
  


  let closeprice = thiskline.data.full_klines[index].close //收盘价
  
  ctx.save()
  //下半部分
  ctx.beginPath()
  let down_list = cm_result.y.filter((v:number)=>v<closeprice)
  down_list.forEach((v:any, index:number)=>{
      ctx.lineTo(
        cm_result.x[index] * x_scale,
        (y_max - v) / y_scale
      )      
  })
  ctx.lineTo(0, (y_max - down_list[down_list.length - 1]) / y_scale)
  ctx.lineTo(0, (y_max - down_list[0]) / y_scale)
  // ctx.lineTo(cm_result.x[0] * x_scale, (y_max - down_list[0]) / y_scale)
  var linear = ctx.createLinearGradient(0, 0, 250, 0);
  linear.addColorStop(0.0, '#F0927D');
  linear.addColorStop(1.0, '#FCE6DF');
  ctx.fillStyle = linear
  ctx.fill()


  //上半部分
  ctx.beginPath()
  let up_list = cm_result.y.filter((v:number)=>v>=closeprice)
  up_list.forEach((v:any, index:number)=>{
      ctx.lineTo(
        cm_result.x[ down_list.length + index] * x_scale,
        (y_max - v) / y_scale
      )      
  })
  up_list = [down_list[down_list.length - 1]].concat(up_list)
  ctx.lineTo(0, (y_max - up_list[up_list.length - 1]) / y_scale)
  ctx.lineTo(0, (y_max - up_list[0]) / y_scale)
  ctx.lineTo(cm_result.x[ down_list.length ] * x_scale, (y_max - up_list[0]) / y_scale)
  var linear2 = ctx.createLinearGradient(0, 0, 250, 0);
  linear2.addColorStop(0.0, '#88B4FB');
  linear2.addColorStop(1.0, '#C4E2FF');
  ctx.fillStyle = linear2
  ctx.fill()
  ctx.restore()
  

  //平均成本线
  let average_index = 0 //与平均成本最接近的序号
  ctx.save()
  cm_result.x.forEach((v:any, index:number)=>{
    if(Math.abs(cm_result.y[index] - cm_result.avgCost) <= Math.abs(cm_result.y[average_index] - cm_result.avgCost)) average_index = index
  })
  if(average_index > 0){
    ctx.save()
    ctx.beginPath()
    ctx.moveTo(0, (y_max - cm_result.avgCost) / y_scale)
    ctx.lineTo(cm_result.x[average_index] * x_scale - 3, (y_max - cm_result.avgCost) / y_scale)
    ctx.lineWidth = 2
    ctx.strokeStyle = '#F97400'
    ctx.setLineDash([6, 2])
    ctx.stroke()
    ctx.restore()
    
    ctx.beginPath();
    ctx.lineWidth = 1.5
    ctx.strokeStyle = '#F97400'
    ctx.arc(cm_result.x[average_index] * x_scale, (y_max - cm_result.avgCost) / y_scale, 3.5, 0, Math.PI * 2, true);  // 右眼
    ctx.stroke();

  }
  ctx.restore()

  //绘制右侧坐标系
  ctx.strokeStyle = thiskline.options.split_line_color
  ctx.font = thiskline.options.font
  thiskline.common_data.price_list.forEach((v,index)=>{
    let y = axisIntAdd((y_max - v) / y_scale)

    let font_y = y + Math.round(thiskline.common_data.font_number_height / 2)

    if(index == 0){
      font_y -= Math.round(thiskline.common_data.font_number_height)
    }
    else if(index == thiskline.common_data.price_list.length - 1){
      font_y += Math.round(thiskline.common_data.font_number_height * 0.5)
    }
    ctx.textAlign = 'right'
    ctx.fillText(
      v.toFixed(thiskline.data.decimal),
      width,
      font_y
    )
    ctx.beginPath()
    ctx.moveTo(0, y)
    ctx.lineTo(width, y)
    ctx.stroke()
  })

  //平均成本
  ctx.save()
  ctx.textAlign = 'right'
  ctx.font = 'bold 110% sans-serif'
  ctx.fillText(
    cm_result.avgCost,//
    width,
    axisIntAdd((y_max - cm_result.avgCost) / y_scale + thiskline.common_data.font_number_height / 2 ) 
  )  
  ctx.restore()

  // ctx.textAlign = 'right'
  // ctx.fillText(
  //   cm_result.avgCost.toFixed(thiskline.data.decimal),
  //   width,
  //   axisIntAdd((y_max - cm_result.avgCost) / y_scale) + Math.round(thiskline.common_data.font_number_height / 2)
  // )  

  cyq_info_html!.innerHTML = `
    <table class="quotechart2022_c_cyq_info_table">
      <tr>
        <td>日期: </td>
        <td class="qcyq_t_v">${thiskline.data.full_klines[index].date}</td>
      </tr>
      <tr>
        <td>获利比例:</td>
        <td class="qcyq_t_v">${(cm_result.benefitPart * 100).toFixed(2) + '%'}</td>
      </tr>
      <tr>
        <td colspan="2">${blTable(cm_result.benefitPart)}</td>
      </tr>
      <tr>
        <td>平均成本:</td>
        <td class="qcyq_t_v">${cm_result.avgCost}</td>
      </tr>
      <tr>
        <td>90%成本:</td>
        <td class="qcyq_t_v">${cm_result.percentChips['90'].priceRange.join('-')}</td>
      </tr>
      <tr>
        <td>集中度:</td>
        <td class="qcyq_t_v">${(cm_result.percentChips['90'].concentration * 100).toFixed(2) + '%'}</td>
      </tr>
      <tr>
        <td>70%成本:</td>
        <td class="qcyq_t_v">${cm_result.percentChips['70'].priceRange.join('-')}</td>
      </tr>
      <tr>
        <td>集中度:</td>
        <td class="qcyq_t_v">${(cm_result.percentChips['70'].concentration * 100).toFixed(2) + '%'}</td>
      </tr>
    </table>
  `

  let bltd1 = cyq_info_html.querySelector('.bltd1') as HTMLTableCellElement
  let bltd1_span = bltd1.querySelector('span') as HTMLSpanElement
  let bltd2 = cyq_info_html.querySelector('.bltd2') as HTMLTableCellElement
  let bltd2_span = bltd2.querySelector('span') as HTMLSpanElement

  if(bltd1_span.clientWidth > bltd1.clientWidth ){
    bltd1_span.innerText = ''
  }
  if(bltd2_span.clientWidth > bltd2.clientWidth ){
    bltd2_span.innerText = ''
  }  

  //       <tr>
      //   <td>筹码分布</td>
      //   <td></td>
      // </tr>
}