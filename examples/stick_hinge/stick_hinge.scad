

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
	renderType = "obj"; //sizer | obj
stickWallSize = 2;
stickRadius = 1.3;
outerBoxWidth = 7;

stickEndWallSize = 2;
holderLength = 15;
partType = "all"; // "first" | "second" | "all"
partShiftApart = 8;

hingeGapWidth = 1.5;
hingeGapSize = [10,8,hingeGapWidth];
hingeRadius = 3;
hingeHeight = 4;
hingeOffset = holderLength*0.7;
innerRadiusOffset = 2;

hingeBumpHeight = 1.6;
bumpShrink = 0.15;
maleXShrink = 0.82;
bumpRounding = 0.7;
isLocked = "true";
lockedAngle = 30;



module ring(){
    difference(){
        cyl(r=outerBoxWidth, h=1);
        cyl(r=outerBoxWidth/2.3, h=2);
    }
}

module stickHole(){
cyl(r=stickRadius, h=holderLength);
}


    module stick_holder(hingePart="female", isLocked="false", lockedAngle=0){
        difference(){
            //cyl(r=stickRadius+stickWallSize, h=holderLength, r1=stickRadius*0.9);
            cuboid([outerBoxWidth, outerBoxWidth, holderLength], rounding=1.5);
            up(stickEndWallSize)
            stickHole()
            
            ring();
        }
        if(isLocked == "true"){
        
        } else if(hingePart == "female" ){
        down(hingeOffset)
        rotate([90,0,0])
        difference(){
        union(){
            cyl(r=hingeRadius-0.5, h=hingeHeight);
                
            back(hingeRadius/2+1)
            cuboid([hingeRadius*1.2,hingeRadius*0.9,hingeHeight]);
            }
            cyl(r=hingeRadius-innerRadiusOffset, h=hingeHeight+1);
            cuboid(hingeGapSize);
        }
        }else {
            down(hingeOffset)
            rotate([90,0,0]){
            intersection(){
            union(){
            difference(){
                cyl(r=hingeRadius*0.9, h=hingeHeight/3);
            }
            back(hingeRadius/2+1)
            cuboid([hingeRadius*1.6,hingeRadius,2]);
            }
            if(isLocked == "true"){
            
                scale([1,1.2,1.2])
                cuboid(hingeGapSize);
            } else {
                
                scale([1,1,maleXShrink])
                cuboid(hingeGapSize);
            
            // hinge middle connector

                }
                
            }
                        cyl(r=hingeRadius-innerRadiusOffset-bumpShrink, h=hingeGapWidth+hingeBumpHeight, rounding=bumpRounding);
            }
        }
    }

	module stick_hinge(partType=partType, isLocked=isLocked, lockedAngle=lockedAngle){
    
        rotate([0,90,90])
    //    back(10)maleXShrink
     //   right(3)
        if(partType == "first"|| partType == "all"){
        
        up(partShiftApart)
        rotate([0, lockedAngle, 0])
        
           #stick_holder(hingePart="female", isLocked=isLocked, lockedAngle=lockedAngle);
        }
      //  down(holderLength*0.7)
        //fwd(holderLength*partShiftApart)
       // right(partShiftApart)
        rotate([90,90,0])
        
        rotate([0, lockedAngle, 0])
        if(partType == "second" || partType == "all"){
            #stick_holder(hingePart="male", isLocked=isLocked, lockedAngle=lockedAngle);
        }
		
	}


    sliced(renderType=renderType) {
    
    if(renderType == "sizer"){
    
    difference(){
        cuboid([outerBoxWidth, outerBoxWidth, 3], rounding=0.1);
        up(stickEndWallSize)
        
        down(2)
        fwd(outerBoxWidth/2+0.3)
        rotate([90,0,0])
        text3d(str(stickEndWallSize), center=true, size=2, h=1);
        
        stickHole();
        
        
        
        }
    }else{
        rotate([0,90,0])
            stick_hinge();
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
   // rotate([0,90,90])
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

