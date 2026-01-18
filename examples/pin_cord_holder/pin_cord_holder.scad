

	include <BOSL2/std.scad>;
include <BOSL2/screws.scad>;
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

holeCount = 6;
    holeDiameter = 7;
    holeDepth = 40;
    holeOffset = 15;
    edgeOffset = 16;
    
    screwDiameter = 4.55;
    screwDepth = 50;

    topHoleDiameter = 10;
    topHoleDepth = 10;
    holderLength = holeCount*(holeDiameter/2+holeOffset)-holeOffset+edgeOffset;
    holderWidth = 40;
    holderSize = [holderWidth, holderLength, 12];
    
    screwHoleOne =  [holderWidth/2-screwDiameter,-holderLength/2+screwDiameter,5];
    screwHoleTwo =  [-holderWidth/2+screwDiameter,-holderLength/2+screwDiameter,5];

    screwHoleThree =  [holderWidth/2-screwDiameter,holderLength/2-screwDiameter,5];
    screwHoleFour =  [-holderWidth/2+screwDiameter,holderLength/2-screwDiameter,5];
    
    partType = "screws";


     module cord_holder(screwType="#8-40", isOutline=true, cordLength=20){
        screw(screwType, head="flat", drive="slot", length=cordLength);
     }

    module cord_hole(cordLength=cordLength){
        
//        left(screwHoleDown)
right(4)
        cyl(h=holeDepth, d=holeDiameter);
        
        down(1)
        right(0)
        rotate([0,-90,0])
        cord_holder(isOutline=true, cordLength=cordLength);
    }

    module screw_attach_hole(){
     down(15)
      screw("#8-54", head="flat undercut",length=20);
//        cyl(h=topHoleDepth, d=topHoleDiameter);
 //       cyl(h=screwDepth, d=screwDiameter);
    }

    module corscrew_cutout(){
        cyl(h=screwDepth, d=screwDiameter);
    }
    

	module pin_cord_holder(){
        difference() {
            
		cuboid(holderSize, rounding=3, anchor=TOP);
        union() {
        rotate([0.0,90])
        //#ydistribute(spacing=10){
            for (i = [0 : holeCount]) {
                
                move([0,holeOffset*i+holeOffset-edgeOffset/4-holderLength/2+2, 0.1])
                cord_hole(cordLength=7);
               
            }
            
            
            move(screwHoleOne)
            screw_attach_hole();
            
            move(screwHoleTwo)
            screw_attach_hole();
            
            move(screwHoleThree)
            screw_attach_hole();
            
            move(screwHoleFour)
            screw_attach_hole();
            
            
        }
        }
	}


    sliced(renderType=renderType) {
        if(partType == "screws"){       
        cord_holder(isOutline=false);
        left(10)
          cord_holder(isOutline=false);
           left(20)
           cord_holder(isOutline=false);
         }
         if(partType == "box"){
           pin_cord_holder();
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

