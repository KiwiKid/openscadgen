
	include <BOSL2/std.scad>;
	include <BOSL2/joiners.scad>;

	$fa = .01;
	$fs = $preview ? 5 : 1;
	$fn = 200;

	/*
	renderType:
	use to print a test slice and confirm sizing before printing:
	 - "horzSlice" - horizontal slices (default)
	 - "vertSlice" - vertical slices
	 - "all" - the whole object
	*/
	renderType = "obj";
    
    obj = "holder";

    nutDiameter = 3;
    nutHolderHeight = 15;
    
    wallDiameter = 4;
    tickerHeight = 7;

    tickerLength = 40;
    tickerWidth = 2;

    tickerStyle = "flat"; // "double"
    
    arrowBodyLength = tickerLength*0.5;
    arrowBodyDepth = 3;
    
    arrowBodyHeight = 12;
    
    triLength = 20;
    
    triWidth = 12;
    
    dovetailWidth =8;
    dovetailHeight =4;
    
    holderWidth = 5;
   
   holderHeight =10;
   holderDepth = 15;
   holderRounding = 1;
   
   arrowBodyOffset = 4;
   
   holderOffset = 2;
    
    module arrow(){
        left(arrowBodyOffset)
        cuboid([arrowBodyLength+10,arrowBodyHeight, arrowBodyDepth], anchor=BOTTOM+LEFT, rounding=3, edges="Z");
        
        right(arrowBodyLength)
        linear_extrude(h=arrowBodyDepth)
        right_triangle([triLength,triWidth]);
        
      //  
      
        mirror([0,1,0])
        right(arrowBodyLength)
        linear_extrude(h=arrowBodyDepth)
        right_triangle([triLength,triWidth]);
    }
    
    module dovetailCon(dovetailSlide=10){
    
    rotate([90,0,-90])
        dovetail(slide=dovetailSlide, h=dovetailHeight, w=dovetailWidth, anchor=LEFT+BOTTOM, radius=0.5, round=true);
        }

	module wheel_ticker(){
    
     if(obj == "holder" || obj == "all"){
        
        left(3.5)
		difference(){
        union(){
            cyl(h=nutHolderHeight, d=nutDiameter+wallDiameter, anchor=BOTTOM);
            
            
             right(holderOffset)
            cuboid([holderWidth, holderHeight, holderDepth],  rounding=holderRounding, anchor=BOTTOM+LEFT);
           }
        
        
      //  up(tickerHeight+tickerHeight/2)
        //    
        //    
       // down(3)
            
            
             right(holderOffset)
             right(wallDiameter+1)
             up(10)
            back(dovetailWidth/2)
            dovetailCon(dovetailSlide=10);
            
            cyl(h=nutHolderHeight+1, d=nutDiameter, anchor=BOTTOM);
            }
        
        arrow();
        }

        if(obj == "ticker" || obj == "all"){
            up(tickerHeight)
            right(nutDiameter/2)
            cuboid([tickerLength, tickerWidth, tickerHeight], anchor=LEFT+BOTTOM);
            
            up(tickerHeight+tickerHeight/2)
            right(nutDiameter/2)
            back(dovetailWidth/2)

            dovetailCon(dovetailSlide=10);
        }
        
	}


    sliced(renderType=renderType) {
        wheel_ticker();
    }
       








	
     
module sliced(
    renderType = "horzSlice",        // "horzSlice", "vertSlice", or "all"
    sliceSize = 1000,
    sliceThickness = 0.2,
    showRawSlices = false,
    horzSlicePos = [-500, -500, 0],
    vertSlicePos = [0, -500, -500]
) {
   
    module horz_slice(raw=false) {
        if (raw) {
            translate(horzSlicePos)
                cube([sliceSize, sliceSize, sliceThickness], center=false);
        } else {
            intersection() {
                children();
                translate(horzSlicePos)
                    cube([sliceSize, sliceSize, sliceThickness], center=false);
            }
        }
    }

    module vert_slice(raw=false) {
        if (raw) {
            translate(vertSlicePos)
                cube([sliceThickness, sliceSize, sliceSize], center=false);
        } else {
            intersection() {
                children();
                translate(vertSlicePos)
                    cube([sliceThickness, sliceSize, sliceSize], center=false);
            }
        }
    }

    if (renderType == "horzSlice") {
        horz_slice(raw=showRawSlices){
            children();
        }
    } else if (renderType == "vertSlice") {
        vert_slice(raw=showRawSlices){
            children();
        }
    } else if (renderType == "all") {
        // show raw slices for reference
        horz_slice(raw=true);
        vert_slice(raw=true);
        // show full object
        children();
    } else {
        // show full object
        children();
    }
}

