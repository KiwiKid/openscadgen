

	include <BOSL2/std.scad>;

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

   pumpNozzleWidth =2;
    airbedNozzleWidth =6;
    
    airbedFillHoleDiameter = 18;

    
       module funnel(tubeDiameter=1, tubeLength = 5, airbedFillHoleDiameter=airbedFillHoleDiameter, airbedFillLength=7){
    
    adaptorHeight = 6;
		cyl(r1=pumpNozzleWidth, r2=airbedNozzleWidth, h=adaptorHeight)

        attach(BOTTOM){

            cyl(r=tubeDiameter, h=tubeLength);
            }
           
           
           up(adaptorHeight/2+tubeLength/2-2)
            cyl(r=airbedNozzleWidth+airbedFillHoleDiameter, h=airbedFillLength);

        }
        
	module pump_adaptor(){
    
        difference(){
        union(){
        tubeDiameter =4.5;
        
        
            funnel(tubeDiameter=tubeDiameter, airbedFillHoleDiameter=0.5, airbedFillLength=4);
            down(7)
            difference(){
                cyl(r=tubeDiameter+0.5,h=14, chamfer=1);
                down(1)
                cyl(r=tubeDiameter,h=12.5);
                
            }
            }
            up(1)
            funnel(tubeDiameter=2, tubeLength=8, airbedFillHoleDiameter=0, airbedFillLength=6.1);
        }
        
	}
    
    
module        pump_adaptor2(){
nozzleRadius = 8.1;

ringRadius = 11;
holderRingSize = 3;
holderRingHeight = 2;
    adaptorHeight = 12;
    
    ringUp = 3;
    
        difference(){
            cyl(r=ringRadius, h=adaptorHeight);
            cyl(r=nozzleRadius,h=12.5);
}
up(ringUp)
difference(){
            #cyl(r=ringRadius+holderRingSize, h=holderRingHeight,chamfer=1);
            cyl(r=ringRadius,h=12.5);
}
            
}


    sliced(renderType=renderType) {
     /*   up(60)
        scale(2)
        pump_adaptor();*/
        
      //  scale(2)
        pump_adaptor2();
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
                cuboid([sliceSize, sliceSize, sliceThickness], anchor=[-1,-1,-1]);
        } else {
            intersection() {
                children();
                translate(horzSlicePos)
                    cuboid([sliceSize, sliceSize, sliceThickness], anchor=[-1,-1,-1]);
            }
        }
    }

    module vert_slice(raw=false) {
        if (raw) {
            translate(vertSlicePos)
                cuboid([sliceThickness, sliceSize, sliceSize], anchor=[-1,-1,-1]);
        } else {
            intersection() {
                children();
                translate(vertSlicePos)
                    cuboid([sliceThickness, sliceSize, sliceSize], anchor=[-1,-1,-1]);
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

