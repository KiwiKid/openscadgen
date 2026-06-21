

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
    
    topType = "base"; // cupped base
    
    holderRadius = 90;
    holderHeight = 100;
    holderWallWidth = 1;
    
    
    baseRingWidth = 4;
    baseRingHeight = 20;
    cutoutHeight = 10;
    
    toiletPaperHeight = 115;
    toiletPaperCount = 2;
    baseOuterRingHeight = toiletPaperHeight*toiletPaperCount;
    outerRightOffset = baseRingWidth;
    
    endChamfer = 5;

    
    module outerCutoutRing(){
                              // down(1)
        difference(){
                cyl(r=holderRadius+outerRightOffset+1, h=cutoutHeight+1, anchor=BOTTOM);
                
        down(0.1)
        up(2)
                         cyl(r=holderRadius+outerRightOffset-baseRingWidth, h=cutoutHeight, anchor=BOTTOM, chamfer2=-endChamfer);
                }
     }

	module basket_holder_extender(topType=topType){
      difference(){
        cyl(r=holderRadius, h=holderHeight);
        
        cyl(r=holderRadius-holderWallWidth, h=holderHeight+3);
        }
        
if(topType == "inverted"){
     /*   difference(){
            cyl(r=holderRadius, h=holderHeight);
            
            cyl(r=holderRadius, h=holderHeight);
        }*/
        } else if(topType == "cupped"){
        
            cyl(r=holderRadius, h=10);
            } 
            
	}


    sliced(renderType=renderType) {
    if(topType == "base"){
    
    difference(){
        cyl(r=holderRadius, h=baseRingHeight, anchor=BOTTOM);

       
        down(0.1)
        cyl(r=holderRadius-baseRingWidth, h=baseRingHeight+20, anchor=BOTTOM);
    }
     difference(){
        #cyl(r=holderRadius+outerRightOffset, h=baseOuterRingHeight, anchor=BOTTOM);
        down(0.1)
        cyl(r=holderRadius+outerRightOffset-baseRingWidth, h=baseOuterRingHeight+1, anchor=BOTTOM, chamfer2=-endChamfer);
        
        // outer cutout ring
        
                outerCutoutRing();
                
         //       up(baseOuterRingHeight)
        //        outerCutoutRing();
    }
            
         }else{
                basket_holder_extender();
        }
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

