/**
 * 初始化价格走势canvas
 * @param {int} showdays 展示的天数
 * @param {Object} trend_series_data 走势数据
 * @param {String} canvas_id canvas id
 * @param {String} promotion_price 促销价格，用于最后特殊的标记
 */
function init(showdays, trend_series_data, canvas_id, promotion_price) {
    let formated_data = trend_series_data;
    let canvas = document.getElementById(canvas_id);
    let canvas_context = canvas.getContext("2d");

    var devicePixelRatio = window.devicePixelRatio || 1;
    var backingStoreRatio = canvas_context.webkitBackingStorePixelRatio || 1;
    var ratio = devicePixelRatio / backingStoreRatio;

    var w = document.getElementById('trend').offsetWidth;
    var h = document.getElementById('trend').offsetHeight;
    canvas.width = w * ratio;
    canvas.height = h * ratio;
    canvas.style.width = w + "px";
    canvas.style.height = h + "px";
    canvas_context.scale(ratio,ratio);

    if (promotion_price) {
        promotion_price /= 100;
    }
    draw(canvas_context, formated_data, showdays,w ,h, promotion_price);
}

function draw(canvas_context, data, days, w, h, promotion_price) {
    let MaxMin, MaxMinIndex;
    let showLastPoint = false;
    MaxMin = getMaxMin(data);
    MaxMinIndex = getMaxMinIndex(data);

    // 优惠价格更低
    if (promotion_price) {
        // 最后一个价格改为优惠价格
        data[data.length-1].y = promotion_price;
        showLastPoint = true;

        MaxMin = getMaxMin(data);
        MaxMinIndex = getMaxMinIndex(data);
    }

    if (MaxMinIndex[1] == (data.length-1) ) {
        showLastPoint = true;
    }

    let Max = MaxMin[0] * 1.05;
    let Min = MaxMin[1] * 0.95;
    if (showLastPoint) {
        Max = MaxMin[0] * 1.1;
        Min = MaxMin[1] * 0.8;
    }

    let deviceScreen = getScreenInfo();
    let height = h * 0.72;
    let width = w * 0.82;
    let startX = w * 0.123;
    let startY = 16;

    let section = (Max - Min) / 5;

    // 一些配置项
    let drawOptions = {
        drawWidth: width,
        drawHeight: height,
        screenWidth: deviceScreen.width,
        screenHeight: deviceScreen.height,
        interval: days,
        startX: startX,
        startY: startY,
        maxValue: MaxMin[0],
        maxIndex: MaxMinIndex[0],
        minValue: MaxMin[1],
        minIndex: MaxMinIndex[1],

        xAxis: {
            font: {
                size: 20,
                color: '#C8C7CD',
                textAlign: 'center'
            },
            line: {
                color: '#EEEEEE',
                width: 1
            }
        },

        yAxis: {
            font: {
                size: 20,
                color: '#C8C7CD',
                textAlign: 'right'
            },
            line: {
                color: '#EEEEEE',
                width: 1,
            },
            maxYAxis: Max,
            minYAxis: Min,
            section: section,
        },
        area: {
            color: 'rgba(49,195,178,0.06)',
            borderColor: 'rgba(49,195,178,0.06)'
        },
        line: {
            strokeStyle: '#31C3B2',
            width: 1,
        },
        promotionPrice: promotion_price
    }

    drawAxis(canvas_context, drawOptions, true, false);
    drawLine(canvas_context, data, drawOptions, true);

    let all_x_txt = getDaysByX(data);
    let xAxis_txt_max = formatDay(all_x_txt[drawOptions.maxIndex]);
    let xAxis_txt_min = formatDay(all_x_txt[drawOptions.minIndex]);
    let xAxis_txt_last = formatDay(all_x_txt[data.length-1]);

    let maxPointOptions = null;
    let minPointOptions = null;
    let lastPointOptions = null;

    let showMaxPoint = true;
    if (MaxMinIndex[0] === MaxMinIndex[1]) {
        showMaxPoint = false;
    }

    if (showMaxPoint) {
        let point_x = (drawOptions.drawWidth / data.length) * MaxMinIndex[0] + startX;
        let point_y = (drawOptions.yAxis.maxYAxis - data[MaxMinIndex[0]].y) / (drawOptions.yAxis.maxYAxis - drawOptions.yAxis.minYAxis) * drawOptions.drawHeight;
        let point_v = "￥" + parseFloat(data[MaxMinIndex[0]].y).toFixed(2);
        let rectHeight = 15;
        let txt_len = canvas_context.measureText(point_v).width;

        let patch_x = 0;
        if (MaxMinIndex[0] == data.length-1) {
            patch_x = 0 - txt_len * 0.4;
        }

        maxPointOptions = {
            pointX: point_x,
            pointY: point_y,
            point_color: 'rgba(242,67,67,1)',
            point_color_end: 'rgba(242,67,67,.3)',
            rectTip: {
                x: point_x - txt_len/2 - getPx(deviceScreen.width, 8) + patch_x,
                y: point_y + startY - rectHeight/2 - getPx(deviceScreen.width, 28),
                w: txt_len + getPx(deviceScreen.width, 16),
                h: rectHeight,
                fillStyle: 'rgba(255,255,255,0.90)',
                strokeStyle: '#FF3434',
                lineWidth: 1,
                fontSize: 20,
                fontColor: '#FF3434'
            },
            tipTxt: {
                x: point_x + txt_len/2 + patch_x,
                y: point_y + startY - getPx(drawOptions.screenWidth, 20),
                txt: point_v,
                txtLen: txt_len,
            },
            xAxisTick: {
                index: drawOptions.maxIndex,
                txt: xAxis_txt_max,
                align: 'center',
                x: point_x,
                y: drawOptions.startY + drawOptions.drawHeight + getPx(drawOptions.screenWidth, 30)
            }
        };
    }

    let showMinPoint = true;
    if (showLastPoint && (MaxMinIndex[1] === (data.length-1))) {
        showMinPoint = false;
    }
    if (showMinPoint) {
        let point_x = (width / data.length) * MaxMinIndex[1] + startX;
        let point_y = ((drawOptions.yAxis.maxYAxis - data[MaxMinIndex[1]].y) / (drawOptions.yAxis.maxYAxis - drawOptions.yAxis.minYAxis))  * drawOptions.drawHeight;
        let point_v = "￥" + parseFloat(data[MaxMinIndex[1]].y).toFixed(2);
        let rectHeight = 15;
        let txt_len = canvas_context.measureText(point_v).width;

        let patch_x = 0;
        if (MaxMinIndex[1] == data.length-1) {
            patch_x = 0 - txt_len * 0.4;
        }

        minPointOptions = {
            pointX: point_x,
            pointY: point_y,
            point_color: 'rgba(49,195,178,1)',
            point_color_end: 'rgba(49,195,178,.3)',
            rectTip: {
                x: point_x - txt_len/2 - getPx(deviceScreen.width, 5) + patch_x,
                y: point_y + startY - rectHeight/2 - getPx(deviceScreen.width, 28),
                w: txt_len + getPx(deviceScreen.width, 10),
                h: rectHeight,
                fillStyle: 'rgba(255,255,255,0.90)',
                strokeStyle: '#3BC065',
                lineWidth: 1,
                fontSize: 20,
                fontColor: '#3BC065'
            },
            tipTxt: {
                x: point_x + txt_len/2 + patch_x,
                y: point_y + startY - getPx(drawOptions.screenWidth, 20),
                txt: point_v,
                txtLen: txt_len,
            },
            xAxisTick: {
                index: MaxMinIndex[1],
                txt: xAxis_txt_min,
                align: 'center',
                x: point_x,
                y: drawOptions.startY + drawOptions.drawHeight + getPx(drawOptions.screenWidth, 30)
            }
        };
    }

    if (showLastPoint) {
        let last_price_txt = '';
        if (promotion_price) {
            last_price_txt = "￥"+parseFloat(promotion_price).toFixed(2);
        } else if (MaxMinIndex[1] == (data.length-1)) {
            last_price_txt = "￥"+parseFloat(MaxMin[1]).toFixed(2);
        }

        let txt_len = canvas_context.measureText(last_price_txt).width;
        let point_x = (width / data.length) * (data.length-1) + startX;
        let point_y = ((Max - data[(data.length-1)].y) / (Max - Min)) * height;

        // 求出最近几天中y最高的，作为最后标记框的y
        let recent_max_y = point_y;
        let recent_day_len = data.length * 0.14;
        let i = 0;
        for (i = 0; i < recent_day_len; i++) {
            let tmp_y = ((Max - data[(data.length- i - 1)].y) / (Max - Min)) * height;
            if (recent_max_y >= tmp_y) {
                recent_max_y = tmp_y;
            }
        }
        --i;

        // 防止最高价和优惠价格标记重叠
        let moveTop = 0;
        // if (Math.abs((data.length - i - 1) - MaxMinIndex[0]) <= 3) {
        //     moveTop = -getPx(deviceScreen.width, 4);
        // }

        let rectHeight = 16;
        let rectWidth = txt_len + getPx(drawOptions.screenWidth, 16);

        let rect_x = point_x - txt_len;
        let rect_y = recent_max_y + startY - rectHeight/2 - getPx(deviceScreen.width, 28) + moveTop;
        let font_x = rect_x + txt_len/2+ getPx(drawOptions.screenWidth, 8);
        let font_y = recent_max_y + startY - getPx(drawOptions.screenWidth, 21) + moveTop;
        if (recent_max_y != point_y) {
            rect_x = point_x - txt_len;
        }

        lastPointOptions = {
            pointX: point_x,
            pointY: point_y,
            point_color: 'rgba(49,195,178,1)',
            point_color_end: 'rgba(49,195,178,.3)',
            txt: last_price_txt,
            font: {
                color: '#FFFFFF',
                align: 'center',
                fontSize: 20,
                x: font_x,
                y: font_y,
            },
            rect: {
                strokeStyle: '#31C3B2',
                rectWidth: rectWidth,
                rectHeight: rectHeight,
                x: rect_x,
                y: rect_y,
            },
            xAxisTick: {
                font: {
                    color: '#31C3B2',
                    size: 20
                },
                align: 'center',
                index: data.length-1,
                txt: xAxis_txt_last,
                x: point_x,
                y: drawOptions.startY + drawOptions.drawHeight + getPx(drawOptions.screenWidth, 30)
            }
        }
    }

    if (showLastPoint) {
        markLastPoint(canvas_context, drawOptions, lastPointOptions);
    }

    let max_changed = false;
    let min_changed = false;
    if (showMaxPoint) {
        let xTickWidth_max = canvas_context.measureText(xAxis_txt_max).width;
        if (showLastPoint) {
            let xTickWidth_last = canvas_context.measureText(xAxis_txt_last).width;

            if ((lastPointOptions.xAxisTick.x >= maxPointOptions.xAxisTick.x && lastPointOptions.xAxisTick.x <= (maxPointOptions.xAxisTick.x+xTickWidth_max))
                ||
                (maxPointOptions.xAxisTick.x >= lastPointOptions.xAxisTick.x && maxPointOptions.xAxisTick.x <= (lastPointOptions.xAxisTick.x+xTickWidth_last))) {
                maxPointOptions.xAxisTick.x = lastPointOptions.xAxisTick.x - xTickWidth_last - 10;
                max_changed = true;
            }
        }
    }

    if (showMinPoint) {
        let xTickWidth_min = canvas_context.measureText(xAxis_txt_min).width;
        if (showLastPoint) {
            let xTickWidth_last = canvas_context.measureText(xAxis_txt_last).width;

            if ((lastPointOptions.xAxisTick.x >= minPointOptions.xAxisTick.x && lastPointOptions.xAxisTick.x <= (minPointOptions.xAxisTick.x+xTickWidth_min))
                ||
                (minPointOptions.xAxisTick.x >= lastPointOptions.xAxisTick.x && minPointOptions.xAxisTick.x <= (lastPointOptions.xAxisTick.x+xTickWidth_last))) {
                minPointOptions.xAxisTick.x = lastPointOptions.xAxisTick.x - xTickWidth_last - 10;
                min_changed = true;
            }
        }

        // 最大最小是否有重叠
        if (showMaxPoint) {
            let xTickWidth_max = canvas_context.measureText(xAxis_txt_max).width;
            if ((maxPointOptions.xAxisTick.x >= minPointOptions.xAxisTick.x && maxPointOptions.xAxisTick.x <= (minPointOptions.xAxisTick.x+xTickWidth_min))
                ||
                (minPointOptions.xAxisTick.x >= maxPointOptions.xAxisTick.x && minPointOptions.xAxisTick.x <= (maxPointOptions.xAxisTick.x+xTickWidth_max))) {
                // 重叠了
                if (max_changed) {
                    minPointOptions.xAxisTick.x = maxPointOptions.xAxisTick.x - xTickWidth_max - 10;
                    min_changed = true;
                } else {
                    if (MaxMinIndex[0] > MaxMinIndex[1]) {
                        if (MaxMinIndex[0] < (data.length-MaxMinIndex[0])) {
                            maxPointOptions.xAxisTick.x = minPointOptions.xAxisTick.x + xTickWidth_min + 10;
                            max_changed = true;
                        } else {
                            minPointOptions.xAxisTick.x = maxPointOptions.xAxisTick.x - xTickWidth_max - 10;
                            min_changed = true;
                        }
                    } else {
                        if (MaxMinIndex[1] < (data.length-MaxMinIndex[1])) {
                            minPointOptions.xAxisTick.x = maxPointOptions.xAxisTick.x + xTickWidth_max + 10;
                            min_changed = true;
                        } else {
                            maxPointOptions.xAxisTick.x = minPointOptions.xAxisTick.x - xTickWidth_min - 10;
                            max_changed = true;
                        }
                    }
                }
            }
        }
    }

    if (showMaxPoint) {
        markPoint(canvas_context, data, drawOptions, maxPointOptions, true, false, true);
    }

    if (showMinPoint) {
        markPoint(canvas_context, data, drawOptions, minPointOptions, true, false, true);
    }
}

/**
 * 画X轴、Y轴
 * @param {*} context Canvas context
 * @param {Object} options 配置项
 * @param {Boolean} showYTickText 是否显示Y轴刻度
 * @param {Boolean} showXTickText 是否显示X轴刻度
 */
function drawAxis(context, options, showYTickText=true, showXTickText=false) {
    let section = options.yAxis.section;

    context.strokeStyle = options.yAxis.line.color;
    context.lineWidth = options.yAxis.line.width;

    let numToFix = 0;
    if (options.yAxis.maxYAxis > 200) {
        numToFix = 0;
    }

    if ((options.yAxis.maxYAxis - options.yAxis.minYAxis) < 10) {
        numToFix = 2;
    }

    if (showYTickText) {
        context.fillStyle = options.yAxis.font.color;
        context.font = options.yAxis.font.size + "px";
        context.textAlign = options.yAxis.font.textAlign;
    }
    context.beginPath();

    let i = 0;
    let lineX = options.startX;
    let lineY = options.startY;
    while (i < 6) {
        context.moveTo(lineX, lineY);
        context.lineTo(lineX + options.drawWidth, lineY);
        if (showYTickText) {
            context.fillText((options.yAxis.maxYAxis - i*section).toFixed(numToFix), lineX - 8, lineY + 3);
        }
        lineY += options.drawHeight / 5;
        i++;
    }

    i = 0;

    context.strokeStyle = options.xAxis.line.color;
    context.lineWidth = options.xAxis.line.width;

    let xAxisValues = null;
    let jumpLength = options.interval / 5;
    if (showXTickText) {
        context.fillStyle = options.xAxis.font.color;
        context.font = options.xAxis.font.size + "px";
        context.textAlign = options.xAxis.font.textAlign;

        xAxisValues = getDays(options.interval);
    }

    lineY = options.startY;

    while ( i < 6) {
        if (showXTickText) {
            let tmp_x = lineX - 2,
                tmp_y = lineY + options.drawHeight + 15;
            if (i == 0) {
                context.fillText(xAxisValues[0], tmp_x, tmp_y);
            } else if (i == 5) {
                context.fillText(xAxisValues[xAxisValues.length-1], tmp_x, tmp_y);
            } else {
                context.fillText(xAxisValues[jumpLength * i], tmp_x, tmp_y);
            }
        }

        lineX += options.drawWidth / 5;
        if (i < 4) {
            context.moveTo(lineX, lineY);
            context.lineTo(lineX, lineY + options.drawHeight);
        }
        i++;
    }

    context.stroke();
    context.closePath();
}

function getScreenInfo() {
    let screen = window.screen;
    return screen;
}

function getPx(deviceWidth,originRpx){
    return parseInt((deviceWidth / 750) * originRpx)
}

function getMaxMinIndex(data) {
    var maxIndex = 0;
    var minIndex = 0;
    var max = 0;
    var min = 99999999
    for (var i = 0;i<data.length;i++){
        if (data[i].y >= max){
            max = data[i].y;
            maxIndex = i;
        }
        if (data[i].y <= min){
            min = data[i].y;
            minIndex = i;
        }
    }
    return [maxIndex,minIndex]
}

/**
 * 走势图中标记一个点
 * @param {*} context canvas 2d context
 * @param {*} data y轴数据值
 * @param {*} options 全局配置项
 * @param {*} point 要标记的点
 * @param {*} showTip 是否显示价格点虚线
 * @param {*} showTipTxt 是否显示价格标签
 * @param {*} showXAxis 是否显示x轴标签
 */
function markPoint(context, data, options, point, showTip=true, showTipTxt=false, showXAxis=true) {
    // 标点
    context.beginPath();
    var grd = context.createRadialGradient(point.pointX, point.pointY + options.startY, 1.5, point.pointX, point.pointY + options.startY, 5);
    grd.addColorStop(0, point.point_color);
    grd.addColorStop(0.2, point.point_color_end);
    grd.addColorStop(1, 'rgba(255,255,255,.78)');
    context.fillStyle = grd;
    context.fillRect(point.pointX - 4.5, point.pointY + options.startY - 4.5, 9, 9);
    context.stroke();
    context.closePath();

    // 画tip
    if (showTip) {
        context.beginPath();
        context.strokeStyle = point.rectTip.strokeStyle;
        context.setLineDash([2, 2], 0);
        context.lineWidth = 1;
        context.moveTo(point.pointX, point.pointY + options.startY);
        context.lineTo(point.pointX, options.startY + options.drawHeight);
        context.stroke();
        context.closePath();
    }

    if (showTipTxt) {
        context.beginPath();
        context.setLineDash([]);
        context.lineWidth = point.rectTip.lineWidth;
        context.font = point.rectTip.fontSize + 'px';
        context.strokeStyle = point.rectTip.strokeStyle;
        context.fillStyle = point.rectTip.fillStyle;

        context.strokeRect(point.rectTip.x, point.rectTip.y, point.rectTip.w, point.rectTip.h);
        context.fillRect(point.rectTip.x, point.rectTip.y, point.rectTip.w, point.rectTip.h);
        context.closePath();

        context.beginPath();
        context.fillStyle = point.rectTip.fontColor;
        context.fillText(point.tipTxt.txt, point.tipTxt.x, point.tipTxt.y);
    }

    // 画X轴刻度
    if (showXAxis) {
        context.beginPath();
        context.textAlign = point.xAxisTick.align;
        context.fillStyle = point.rectTip.fontColor;
        context.fillText(point.xAxisTick.txt, point.xAxisTick.x, point.xAxisTick.y);
        context.closePath();
    }
}

/**
 * 画走势线和面积图
 * @param {*} context Canvas context
 * @param {Array|Object} data 走势数据
 * @param {Object} options 配置项
 * @param {Boolean} showArea 是否画面积图
 */
function drawLine(context, data, options, showArea = true) {
    context.strokeStyle = options.line.strokeStyle;
    context.lineWidth = options.line.width;
    context.beginPath();

    let i = 0;
    let tmp_x = null,
        tmp_y = null;
    let tmp_width = options.drawWidth / data.length;
    while(i < data.length) {
        tmp_x = options.startX + tmp_width * (i);
        tmp_y = (options.yAxis.maxYAxis - data[i].y) / (options.yAxis.maxYAxis - options.yAxis.minYAxis) * options.drawHeight
        if (i === 0) {
            context.moveTo(tmp_x, tmp_y + options.startY);
        } else {
            context.lineTo(tmp_x, tmp_y + options.startY);
        }
        ++i;
    }
    context.lineTo(options.startX + options.drawWidth, tmp_y + options.startY);
    context.stroke();
    context.closePath();

    if (showArea) {
        i = 0;
        context.lineWidth = 0.1;
        context.strokeStyle = options.area.borderColor;
        context.fillStyle = options.area.color;
        context.beginPath();

        while (i < data.length) {
            tmp_x = tmp_width * i + options.startX;
            tmp_y = ((options.yAxis.maxYAxis - data[i].y) / (options.yAxis.maxYAxis - options.yAxis.minYAxis)) * options.drawHeight;
            if(i === 0){
                context.moveTo(options.startX, options.drawHeight + options.startY);
                context.lineTo(tmp_x, tmp_y + options.startY);
            }else{
                context.lineTo(tmp_x, tmp_y + options.startY);
            }
            ++i;
        }

        context.lineTo(options.startX + options.drawWidth, tmp_y + options.startY);
        context.lineTo(options.startX + options.drawWidth, options.startY + options.drawHeight);
        context.fill();
        context.stroke();
        context.closePath();
    }
}

/**
 * 标记最后一个点
 * @param {*} context Canvas context
 * @param {Object} options 配置项
 * @param {Object} point point的配置项
 */
function markLastPoint(context, options, point) {

    // 标点
    context.beginPath();
    var grd = context.createRadialGradient(point.pointX, point.pointY + options.startY, 1.5, point.pointX, point.pointY + options.startY, 4.5);
    grd.addColorStop(0, point.point_color);
    grd.addColorStop(.2, point.point_color_end);
    grd.addColorStop(1, 'rgba(255,255,255,.78)');
    context.fillStyle = grd;
    context.fillRect(point.pointX - 4.5, point.pointY + options.startY - 4.5, 9, 9);
    context.stroke();
    context.closePath();

    // 画线
    context.beginPath();
    context.setLineDash([2, 2], 0);
    context.strokeStyle = point.rect.strokeStyle;
    context.moveTo(point.pointX + 0.5, point.pointY + options.startY);
    context.lineTo(point.pointX + 0.5, options.startY + options.drawHeight);
    context.stroke();
    context.setLineDash([]);
    context.closePath();

    // 画Tip
    let txt_len = context.measureText(point.txt).width;
    context.beginPath();
    context.strokeStyle = point.rect.strokeStyle;
    context.fillStyle = point.rect.strokeStyle;
    context.fillRect(point.rect.x, point.rect.y, point.rect.rectWidth, point.rect.rectHeight);
    context.closePath();

    context.beginPath();
    context.font = point.font.fontSize + "px";
    context.textAlign = point.font.align;
    context.fillStyle = point.font.color;
    context.fillText(point.txt, point.font.x, point.font.y);
    context.closePath();

    // 画x轴
    context.beginPath();
    context.textAlign = point.xAxisTick.align;
    context.fillStyle = point.xAxisTick.font.color;
    context.fillText(point.xAxisTick.txt, point.xAxisTick.x, point.xAxisTick.y);
    context.closePath();
}

function getDaysByX(data) {
    let days = []
    data.forEach(p => {
        let d = new Date(p.x * 1e3),
            d_m = d.getMonth() + 1;
        d_m = d_m < 10 ? '0'+d_m : d_m;
        days.push(d_m + '-' + d.getDate())
    });

    return days
}

//读取日期
function getDays(interval) {
    var days = [];
    for(var i = interval-1;i >= 0;i--){
        var time = new Date(new Date().toLocaleDateString()).getTime()  - 86400000 * i
        var ThatDate = new Date(time)
        var month = ThatDate.getMonth() + 1
        month = month < 10 ? '0'+month : month;
        var day = ThatDate.getDate()
        days.push(month +'-'+day);
    }
    return days
}

function getMaxMin(d){
    //最小值
    Array.prototype.min = function() {
        var min = this[0].y;
        var len = this.length;
        for (var i = 1; i < len; i++){
            if (this[i].y < min){
                min = this[i].y;
            }
        }
        return min;
    }
    Array.prototype.max = function() {
        var max = this[0].y;
        var len = this.length;
        for (var i = 1; i < len; i++){
            if (this[i].y > max) {
                max = this[i].y;
            }
        }
        return max;
    }
    return [d.max(),d.min()]
}

function formatData(interval,data) {
    data = data['data'];

    let patched_data = [],
        prev_point = data[0],
        k = 1;
    var current_date = new Date(new Date().toLocaleDateString()).getTime() / 1000;

    prev_point['y'] /= 100
    patched_data.push(prev_point)
    while (k < data.length) {
        if (data[k] && (data[k]['x'] == (prev_point['x'] + 86400))) {
            prev_point = data[k]
            prev_point['y'] /= 100
            k++
        } else {
            prev_point['x'] += 86400
        }

        patched_data.push(prev_point)
    }

    return patched_data;
}

function formatDay(day) {
    let day_arr = day.split('-');
    if (day_arr.length == 2) {
        if (parseInt(day_arr[1]) < 10) {
            day_arr[1] = '0'+day_arr[1];
        }
    }
    return day_arr.join('-');
}