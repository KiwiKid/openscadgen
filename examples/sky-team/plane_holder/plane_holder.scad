

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
    
    partType = "all"; // "rail"

    
    holderWidth = 46.01;
    holderHeight = 8;

  innerWidth = 38;
    innerLength = 29;
    
    bankOffset = -66;
    
    innerHeight = 3;
    boxCount = 8;
    wallWidth = 3;
    
    outerRounding = 2;
    
    slotRotate = 0.5;
    slotDown = 2.35;
    slotRailRouning = 2;

    
    cardDown = 3.1;
    cardWidth = 44;
    cardHeight = 100;
    cardDepth = 2;
    
                 slotWidth  =10;

    
    
    lipUp = 115.5;
    lipDepress = -0.2;
    
    
topLipUp =  9.5;
    lipInnerOffset = 12  ;
    
    hookDrop = 2;
    
    lipOuterSize = [11,holderWidth+8,4];
    lipInnerSize = [33,holderWidth+6,10];
    
        topBoxWidth = innerWidth;
        
        connectorSize = [3, 47, 5];

        
       connectorMove = [-10, 0, 4];
doveHeight = 8;
doveWidth = 10;
doveThick = 4;

innerLipDown = 4;

    topBoxLength = 12;
    
    bottomTabLength = 3;
    bottomTabScale = [5,1,1];
    
    module holderBox(width=50, length=28, height=8, wallWidth=3, rounding=0.2){
    
    difference(){
            
                cuboid([length,width,height], rounding=rounding, edges=TOP);
                cuboid([length-wallWidth,width-wallWidth,height+1]);
                
                }
                }

	module plane_holder(){
        
        
        
//        up(2)
        if(partType == "slot" || partType == "all"){
        
        
                     // lip
                      
             
            right((innerLength-wallWidth/2)*4)
            difference(){
            
            union(){
             holderBox(width=holderWidth, length=innerLength, wallWidth=wallWidth, height=holderHeight,  rounding=-1);
             
          // Bridge slot connector (front)
          move(connectorMove)
         cuboid(connectorSize, rounding=1);
         
     //    rotate([90,0,0])
    //     move(connectorMove)
         //#dovetail("male", thickness=doveThick, h=doveHeight, w=doveWidth);
         
         
         
         /*
         
         
         move(connectorMove)
         cuboid(connectorSize);
         
         rotate([90,0,0])
         move(connectorMove)
         #dovetail("female", thickness=doveThick, h=doveHeight, w=doveWidth);*/
        
    
         
         // brige connector
        /* up(2)
         left(innerLength/2-2)
         cuboid([4,holderWidth,10]);*/
         
             
             }
             
             

             
             down(slotDown)
             rotate([0, slotRotate,0])
            
            holderBox(width=innerWidth*1.07, length=innerLength+3, wallWidth=20, height=innerHeight*1.05, rounding=-slotRailRouning);
             
            // down(cardDown)
             // cuboid([cardHeight,cardWidth,cardDepth]);
                             
                             
               // make into U-shape
               left(10)
            cuboid([innerLength,innerWidth+5,10]);

             }
         }
         
         
        if(partType == "rail" || partType == "all"){
            xcopies(innerLength-wallWidth/2, n=boxCount){
                holderBox(width=innerWidth, length=innerLength, wallWidth=wallWidth, height=innerHeight, rounding=-slotRailRouning  );
                
            }
            
            //top box
            left(lipUp)
            holderBox(width=topBoxWidth, length=topBoxLength, wallWidth=wallWidth, height=innerHeight);
            
            // top lip
            left(lipUp+topBoxLength-wallWidth/2-topLipUp)
            fwd(7 )
            up(-1.5)
            difference(){
            
            down(lipDepress)
                cuboid(lipOuterSize, rounding=0.2);
                
                up(innerLipDown)
                right(lipInnerOffset)
                cuboid(lipInnerSize, rounding=0.2);
            }
            
          // bottom tab
            right(lipUp+topBoxLength-wallWidth/2-11+bottomTabLength)
             rotate([0, slotRotate,0])
             scale(bottomTabScale)
             cuboid([bottomTabLength, innerWidth, innerHeight], chamfer=1, edges=RIGHT);
        }
        
        if(partType == "slot" || partType == "all"){
        
            
            // bank connector
            left(bankOffset-29)
            fwd(28)
            cuboid([2,12,8], rounding=0.5);
            
            
            // top bottom box
            left(bankOffset)
            fwd(innerWidth+14)
            holderBox(width=innerWidth, length=60, wallWidth=wallWidth, height=holderHeight, rounding=0.2);
            
            // top bank box
            left(bankOffset+45)
            fwd(innerWidth+14)
            holderBox(width=innerWidth, length=30, wallWidth=wallWidth, height=holderHeight, rounding=0.2);
            
            // lip to hold up slot
             /*up(1)
             right(lipUp+topBoxLength-wallWidth/2-1)
             back(3)
             rotate([0, slotRotate,0])
             cuboid([4, innerWidth*1.16, 1], rounding=0.5);
             */
             
            // lip to hold up slot (side)
             up(-1+slotDown)
             right(lipUp+topBoxLength-wallWidth/2-16)
             back(innerWidth/2+3)
             
             rotate([0, slotRotate,0])
             union(){
             cuboid([30, 2, 5], rounding=-1.2, edges=TOP-LEFT);
             
             

             rotate([0, slotRotate,0])
             cuboid([30, 2, 5], rounding=-1.2, edges=BOTTOM+BACK);
            }
            
        }
	}
    


    sliced(renderType=renderType) {
        plane_holder();
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

